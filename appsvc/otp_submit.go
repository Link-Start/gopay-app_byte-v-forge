package appsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type otpResumePayload struct {
	JobID             string         `json:"job_id,omitempty"`
	AccountID         string         `json:"account_id,omitempty"`
	N8NExecutionID    string         `json:"n8n_execution_id,omitempty"`
	Action            string         `json:"action,omitempty"`
	Step              string         `json:"step,omitempty"`
	Channel           string         `json:"channel"`
	Target            string         `json:"target"`
	OTP               string         `json:"otp"`
	OTPSource         string         `json:"otp_source"`
	OTPReceivedAtUnix int64          `json:"otp_received_at_unix"`
	Data              map[string]any `json:"data,omitempty"`
}

func (h gopayHTTPHandler) handleOTPSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload otpSubmitPayload
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid otp submit payload"))
		return
	}
	submitted, err := h.submitChannelOTP(r.Context(), payload)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusAccepted, submitted)
}

func (h gopayHTTPHandler) submitChannelOTP(ctx context.Context, payload otpSubmitPayload) (map[string]any, error) {
	channel := normalizeActionOTPChannel(payload.Channel)
	target := normalizeChannelOTPTarget(channel, payload.Target)
	code := normalizeOTP(payload.OTP)
	if channel == "" || target == "" || code == "" {
		return nil, fmt.Errorf("channel, target and otp are required")
	}
	receivedAt := time.Now().Unix()
	if err := h.service.store.SaveLatestChannelOTP(ctx, latestChannelOTP{Channel: channel, Target: target, OTP: code, ReceivedAtUnix: receivedAt, Source: channel}, 30*time.Minute); err != nil {
		return nil, err
	}
	waits, err := h.service.store.PendingChannelOTPWaits(ctx, channel, target, receivedAt)
	if err != nil {
		return nil, err
	}
	resumed := make([]string, 0, len(waits))
	for _, wait := range waits {
		claimed, err := h.service.store.ClaimChannelOTPWait(ctx, wait.JobID, time.Minute)
		if err != nil {
			return nil, err
		}
		if !claimed {
			continue
		}
		if err := postOTPResume(ctx, wait, code, receivedAt); err != nil {
			_ = h.service.store.ReleaseChannelOTPWaitClaim(ctx, wait.JobID)
			return nil, err
		}
		_ = h.service.store.DeleteChannelOTPWait(ctx, wait)
		resumed = append(resumed, wait.JobID)
	}
	return map[string]any{
		"success":              true,
		"channel":              channel,
		"target":               target,
		"otp_received_at_unix": receivedAt,
		"resume_count":         len(resumed),
		"resumed_job_ids":      resumed,
	}, nil
}

func postOTPResume(ctx context.Context, wait channelOTPWaitEntry, code string, receivedAt int64) error {
	body := otpResumePayload{
		JobID:             wait.JobID,
		AccountID:         wait.AccountID,
		N8NExecutionID:    wait.N8NExecutionID,
		Action:            wait.Action,
		Step:              wait.StepName,
		Channel:           wait.Channel,
		Target:            wait.Target,
		OTP:               code,
		OTPSource:         firstNonEmpty(wait.Channel, "channel"),
		OTPReceivedAtUnix: receivedAt,
		Data: map[string]any{
			"otp_found":             true,
			"otp_channel":           wait.Channel,
			"channel_otp_target":    wait.Target,
			"otp_issued_after_unix": wait.IssuedAfterUnix,
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(wait.ResumeURL), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("resume channel otp wait: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("resume channel otp wait returned HTTP %d", resp.StatusCode)
	}
	return nil
}
