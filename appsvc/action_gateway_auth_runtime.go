package appsvc

import (
	"context"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) startAuth(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.AuthStart(ctx, &pb.AuthStartRequest{Phone: stringx.FirstNonEmpty(req.Phone, req.WAPhone), CountryCode: req.CountryCode, Pin: req.Pin, OtpChannel: req.OTPChannel, StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "login", resp.GetSuccess(), nil)
	out.applyOTP(resp.GetOtpSent(), resp.GetVerificationMethod())
	out.StateJSON = resp.GetStateJson()
	out.Phone = stringx.FirstNonEmpty(req.Phone, req.WAPhone)
	out.Ready = resp.GetReady()
	out.AccountTokenReady = resp.GetReady()
	out.SignupPINComplete = !resp.GetPinSetupRequired()
	out.ErrorMessage = resp.GetErrorMessage()
	if resp.GetReady() || resp.GetOtpSent() {
		_ = h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
	}
	return out, nil
}

func (h gopayHTTPHandler) completeAuth(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.AuthComplete(ctx, &pb.AuthCompleteRequest{Otp: req.otpValue(), Pin: req.Pin, StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "login", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.Phone = resp.GetPhone()
	out.Ready = resp.GetReady()
	out.AccountTokenReady = resp.GetReady()
	out.SignupPINComplete = resp.GetPinSetupComplete()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfReady(ctx, req.withOTPChannel(out.OTPChannel), out)
	return out, nil
}
