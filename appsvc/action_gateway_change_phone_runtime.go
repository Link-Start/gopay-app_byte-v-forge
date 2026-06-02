package appsvc

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) acquireChangePhone(req gopayActionRequest) *gopayActionResult {
	out := h.baseResult(req, "change_phone_get_number", strings.TrimSpace(req.Phone) != "", nil)
	out.Phone = req.Phone
	out.ActivationID = req.ActivationID
	if out.Phone == "" {
		out.ErrorMessage = "change phone target phone is required"
	}
	return out
}

func (h gopayHTTPHandler) startChangePhone(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.ChangePhoneStart(ctx, &pb.ChangePhoneStartRequest{Pin: req.Pin, NewPhone: req.Phone, CountryCode: req.CountryCode, StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "change_phone_start", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.Phone = resp.GetNewPhone()
	out.ActivationID = resp.GetActivationId()
	out.applyOTP(resp.GetOtpSent(), "sms")
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) retryChangePhone(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.ChangePhoneRetry(ctx, &pb.ChangePhoneRetryRequest{StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "change_phone_retry", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.applyOTP(resp.GetOtpSent(), "sms")
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) completeChangePhone(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.ChangePhoneComplete(ctx, &pb.ChangePhoneCompleteRequest{Otp: req.otpValue(), StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "change_phone_complete", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.ChangePhoneComplete = resp.GetSuccess()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfAccount(ctx, req, out.StateJSON)
	return out, nil
}
