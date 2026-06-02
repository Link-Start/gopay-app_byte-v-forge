package appsvc

import (
	"context"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) finish(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	if err := h.saveStateIfAccount(ctx, req, req.StateJSON); err != nil {
		return nil, err
	}
	job, _ := h.loadWorkflowJob(ctx, req.JobID)
	if job != nil {
		job.Status = "succeeded"
		job.Result = protoStruct(req.Data)
		job.UpdatedAtUnix = time.Now().Unix()
		_ = h.saveWorkflowJob(ctx, job)
	}
	return h.baseResult(req, "finish", true, req.Data), nil
}

func (h gopayHTTPHandler) fail(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	job, _ := h.loadWorkflowJob(ctx, req.JobID)
	if job != nil {
		job.Status = "failed"
		job.ErrorMessage = req.ErrorMessage
		job.Result = protoStruct(req.Data)
		job.UpdatedAtUnix = time.Now().Unix()
		_ = h.saveWorkflowJob(ctx, job)
	}
	out := h.baseResult(req, "fail", false, req.Data)
	out.ErrorMessage = req.ErrorMessage
	return out, nil
}

func (h gopayHTTPHandler) genericSuccess(req gopayActionRequest, step string, data map[string]any) *gopayActionResult {
	return h.baseResult(req, step, true, data)
}

func (h gopayHTTPHandler) baseResult(req gopayActionRequest, step string, success bool, data map[string]any) *gopayActionResult {
	return &gopayActionResult{JobID: req.JobID, N8NExecutionID: req.N8NExecutionID, Action: stringx.FirstNonEmpty(req.actionScope, gopayAccountActionScope), Step: step, Operation: normalizeGoPayWorkflowOperation(req.Operation), Success: success, GopayAccountID: req.GopayAccountID, Phone: req.Phone, OTPChannel: normalizeActionOTPChannel(req.OTPChannel), CountryCode: req.CountryCode, ActivationID: req.ActivationID, FailureCount: req.FailureCount, StateJSON: stringx.FirstNonEmpty(req.StateJSON, "{}"), Data: data}
}

func (h gopayHTTPHandler) basePaymentResult(scope string, req gopayActionRequest, step string, success bool, data map[string]any) *gopayActionResult {
	return &gopayActionResult{JobID: req.JobID, N8NExecutionID: req.N8NExecutionID, Action: strings.TrimSpace(scope), Step: step, Success: success, GopayAccountID: req.GopayAccountID, Phone: stringx.FirstNonEmpty(req.Phone, req.WAPhone), OTPChannel: normalizeActionOTPChannel(stringx.FirstNonEmpty(req.OTPChannel, req.Channel)), CountryCode: req.CountryCode, ActivationID: req.ActivationID, StateJSON: stringx.FirstNonEmpty(req.StateJSON, "{}"), Data: data}
}

func (r *gopayActionRequest) withJob(job *pb.GopayWorkflowJob) gopayActionRequest {
	if job == nil {
		r.Operation = normalizeGoPayWorkflowOperation(r.Operation)
		r.OTPChannel = normalizeActionOTPChannel(r.OTPChannel)
		return *r
	}
	r.JobID = stringx.FirstNonEmpty(r.JobID, job.GetJobId())
	r.N8NExecutionID = stringx.FirstNonEmpty(r.N8NExecutionID, job.GetN8NExecutionId())
	r.Operation = normalizeGoPayWorkflowOperation(stringx.FirstNonEmpty(r.Operation, job.GetOperation()))
	r.GopayAccountID = goPayAppAccountID(stringx.FirstNonEmpty(r.GopayAccountID, job.GetGopayAccountId()))
	r.Phone = stringx.FirstNonEmpty(r.Phone, job.GetPhone())
	r.CountryCode = stringx.FirstNonEmpty(r.CountryCode, job.GetCountryCode())
	r.Pin = stringx.FirstNonEmpty(r.Pin, job.GetPin())
	r.OTPChannel = normalizeActionOTPChannel(stringx.FirstNonEmpty(r.OTPChannel, job.GetOtpChannel()))
	return *r
}

func (r gopayActionRequest) withOTPChannel(channel string) gopayActionRequest {
	if normalized := normalizeActionOTPChannel(channel); normalized != "" {
		r.OTPChannel = normalized
	}
	return r
}

func (r gopayActionRequest) otpValue() string {
	for _, value := range []string{r.OTP, mapString(r.Data, "otp"), mapString(r.Data, "channel_otp"), mapString(r.Data, "code")} {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func (r *gopayActionResult) applyOTP(sent bool, method string) {
	r.OTPSent = sent
	r.OTPRequired = sent
	r.OTPIssuedAfterUnix = time.Now().Unix()
	r.OTPTimeoutSeconds = 300
	if channel := normalizeActionOTPChannel(method); channel != "" {
		r.OTPChannel = channel
	}
}

func (h gopayHTTPHandler) saveStateIfReady(ctx context.Context, req gopayActionRequest, out *gopayActionResult) error {
	if out == nil || (!out.Ready && !out.AccountTokenReady && !out.SignupPINComplete && !out.SignupComplete) {
		return nil
	}
	return h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
}

func (h gopayHTTPHandler) saveStateIfAccount(ctx context.Context, req gopayActionRequest, rawStateJSON string) error {
	if strings.TrimSpace(req.GopayAccountID) == "" || strings.TrimSpace(rawStateJSON) == "" || strings.TrimSpace(rawStateJSON) == "{}" {
		return nil
	}
	state := h.service.parseRequestState(rawStateJSON)
	if channel := normalizeActionOTPChannel(req.OTPChannel); channel != "" {
		state["_otp_channel"] = channel
	}
	persistGopayAccountOTPChannel(state)
	rawStateJSON = stateJSON(state)
	_, err := h.service.SaveGopayAccountState(ctx, &pb.SaveGopayAccountStateRequest{GopayAccountId: req.GopayAccountID, StateJson: rawStateJSON})
	if err != nil {
		return err
	}
	return h.syncAccountProfile(ctx, req, state)
}

func (h gopayHTTPHandler) syncAccountProfile(ctx context.Context, req gopayActionRequest, state stateMap) error {
	return h.syncAccountProfileFields(
		ctx,
		req.GopayAccountID,
		stringx.FirstNonEmpty(req.WAPhone, req.Phone),
		req.CountryCode,
		req.Pin,
		req.OTPChannel,
		state,
	)
}

func (h gopayHTTPHandler) syncAccountProfileFields(ctx context.Context, accountID string, phone string, countryCode string, pin string, otpChannel string, state stateMap) error {
	if strings.TrimSpace(accountID) == "" {
		return nil
	}
	phone = stringx.FirstNonEmpty(
		phone,
		stateString(state, "phone"),
		stateString(state, "_login_phone"),
		stateString(state, "_signup_phone"),
	)
	countryCode = stringx.FirstNonEmpty(countryCode, stateCountryCode(state))
	otpChannel = normalizeActionOTPChannel(stringx.FirstNonEmpty(
		otpChannel,
		gopayAccountOTPChannelFromState(state),
	))
	if strings.TrimSpace(phone) == "" && strings.TrimSpace(countryCode) == "" && strings.TrimSpace(pin) == "" && otpChannel == "" {
		return nil
	}
	_, err := h.service.SaveGopayAccountProfile(ctx, &pb.SaveGopayAccountProfileRequest{
		GopayAccountId: accountID,
		WaPhone:        phone,
		CountryCode:    countryCode,
		OtpChannel:     otpChannel,
		Pin:            pin,
	})
	return err
}
