package appsvc

import (
	"context"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) startPIN(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CreatePinStart(ctx, &pb.CreatePinStartRequest{Pin: req.Pin, OtpChannel: req.OTPChannel, StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "ensure_pin_setup", resp.GetSuccess(), nil)
	out.applyOTP(resp.GetOtpSent(), resp.GetVerificationMethod())
	out.StateJSON = resp.GetStateJson()
	out.ErrorMessage = resp.GetErrorMessage()
	state := h.service.parseRequestState(out.StateJSON)
	ready := resp.GetSuccess() && stateString(state, "stage") == "ready"
	out.SignupPINComplete = ready
	out.AccountTokenReady = ready
	out.Ready = ready
	_ = h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
	return out, nil
}

func (h gopayHTTPHandler) retryPIN(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CreatePinRetry(ctx, &pb.CreatePinRetryRequest{StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "ensure_pin_setup", resp.GetSuccess(), nil)
	out.applyOTP(resp.GetOtpSent(), "")
	out.StateJSON = resp.GetStateJson()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
	return out, nil
}

func (h gopayHTTPHandler) completePIN(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CreatePinComplete(ctx, &pb.CreatePinCompleteRequest{Otp: req.otpValue(), Pin: req.Pin, StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "ensure_pin_setup", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.Phone = resp.GetPhone()
	out.SignupPINComplete = resp.GetPinSetupComplete()
	out.AccountTokenReady = resp.GetPinSetupComplete()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfReady(ctx, req, out)
	return out, nil
}
