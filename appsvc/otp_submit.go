package appsvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) handleOTPSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload pb.SubmitChannelOTPRequest
	if err := protojsonhttp.ReadRequest(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid otp submit payload"))
		return
	}
	submitted, err := h.submitChannelOTP(r.Context(), &payload)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	_ = protojsonhttp.WriteResponse(w, http.StatusAccepted, submitted)
}

func (h gopayHTTPHandler) submitChannelOTP(ctx context.Context, payload *pb.SubmitChannelOTPRequest) (*pb.SubmitChannelOTPResponse, error) {
	channel := normalizeActionOTPChannel(payload.GetChannel())
	target := normalizeChannelOTPTarget(channel, payload.GetTarget())
	accountID := strings.TrimSpace(payload.GetGopayAccountId())
	code := normalizeOTP(payload.GetOtp())
	if code == "" {
		return nil, fmt.Errorf("otp is required")
	}
	receivedAt := time.Now().Unix()
	if payload.GetManualOnce() {
		waits, err := h.pendingManualOTPWaits(ctx, accountID, channel, target, receivedAt)
		if err != nil {
			return nil, err
		}
		resumed, err := h.resumeChannelOTPWaits(ctx, waits, code, receivedAt)
		if err != nil {
			return nil, err
		}
		submitted := &pb.SubmitChannelOTPResponse{
			Success:        len(resumed) > 0,
			ManualOnce:     true,
			GopayAccountId: accountID,
			ResumeCount:    int32(len(resumed)),
			ResumedJobIds:  resumed,
		}
		if len(resumed) == 0 {
			submitted.ErrorMessage = "no pending OTP wait for this GoPay account"
		}
		return submitted, nil
	}
	if channel == "" || target == "" {
		return nil, fmt.Errorf("channel and target are required")
	}
	if err := h.service.store.SaveLatestChannelOTP(ctx, &pb.LatestChannelOTP{Channel: channel, Target: target, Otp: code, ReceivedAtUnix: receivedAt, Source: stringx.FirstNonEmpty(payload.GetOtpSource(), channel)}, 30*time.Minute); err != nil {
		return nil, err
	}
	waits, err := h.service.store.PendingChannelOTPWaits(ctx, channel, target, receivedAt)
	if err != nil {
		return nil, err
	}
	resumed, err := h.resumeChannelOTPWaits(ctx, waits, code, receivedAt)
	if err != nil {
		return nil, err
	}
	return &pb.SubmitChannelOTPResponse{
		Success:           true,
		Channel:           channel,
		Target:            target,
		OtpReceivedAtUnix: receivedAt,
		ResumeCount:       int32(len(resumed)),
		ResumedJobIds:     resumed,
	}, nil
}

func (h gopayHTTPHandler) pendingManualOTPWaits(ctx context.Context, accountID string, channel string, target string, receivedAt int64) ([]*pb.ChannelOTPWaitEntry, error) {
	if accountID != "" {
		waits, err := h.service.store.PendingAccountOTPWaits(ctx, accountID, receivedAt)
		if err != nil || len(waits) > 0 || channel == "" || target == "" {
			return waits, err
		}
	}
	if channel == "" || target == "" {
		return nil, fmt.Errorf("gopay_account_id or channel and target are required")
	}
	return h.service.store.PendingChannelOTPWaits(ctx, channel, target, receivedAt)
}

func (h gopayHTTPHandler) resumeChannelOTPWaits(ctx context.Context, waits []*pb.ChannelOTPWaitEntry, code string, receivedAt int64) ([]string, error) {
	resumed := make([]string, 0, len(waits))
	for _, wait := range waits {
		if wait == nil {
			continue
		}
		claimed, err := h.service.store.ClaimChannelOTPWait(ctx, wait.GetJobId(), time.Minute)
		if err != nil {
			return nil, err
		}
		if !claimed {
			continue
		}
		if err := postOTPResume(ctx, wait, code, receivedAt); err != nil {
			_ = h.service.store.ReleaseChannelOTPWaitClaim(ctx, wait.GetJobId())
			return nil, err
		}
		_ = h.service.store.DeleteChannelOTPWait(ctx, wait)
		resumed = append(resumed, wait.GetJobId())
	}
	return resumed, nil
}

func postOTPResume(ctx context.Context, wait *pb.ChannelOTPWaitEntry, code string, receivedAt int64) error {
	body := &pb.ChannelOTPResumeRequest{
		JobId:             wait.GetJobId(),
		AccountId:         wait.GetAccountId(),
		N8NExecutionId:    wait.GetN8NExecutionId(),
		Action:            wait.GetAction(),
		Step:              wait.GetStepName(),
		Channel:           wait.GetChannel(),
		Target:            wait.GetTarget(),
		Otp:               code,
		OtpSource:         stringx.FirstNonEmpty(wait.GetChannel(), "channel"),
		OtpReceivedAtUnix: receivedAt,
		Data: &pb.ChannelOTPResumeData{
			OtpFound:           true,
			OtpChannel:         wait.GetChannel(),
			ChannelOtpTarget:   wait.GetTarget(),
			OtpIssuedAfterUnix: wait.GetIssuedAfterUnix(),
		},
	}
	if err := postJSON(ctx, wait.GetResumeUrl(), body, jsonPostOptions{
		Timeout:   15 * time.Second,
		Operation: "resume channel otp wait",
	}); err != nil {
		return fmt.Errorf("resume channel otp wait: %w", err)
	}
	return nil
}
