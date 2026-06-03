package appsvc

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) acquireSignupPhone(req gopayActionRequest) *gopayActionResult {
	phone := stringx.FirstNonEmpty(req.Phone, req.WAPhone, mapString(req.Data, "phone"))
	countryCode := stringx.FirstNonEmpty(req.CountryCode, mapString(req.Data, "country_code"), "62")
	data := map[string]any{
		"phone":        phone,
		"country_code": countryCode,
	}
	out := h.baseResult(req, "gopay_app_signup_phone", phone != "", data)
	out.Phone = phone
	out.CountryCode = countryCode
	out.ActivationID = req.ActivationID
	if phone == "" {
		out.ErrorMessage = "signup phone is required"
	}
	return out
}

func (h gopayHTTPHandler) checkSignupPhone(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CheckPhone(ctx, &pb.CheckPhoneRequest{
		Phone:       req.Phone,
		CountryCode: stringx.FirstNonEmpty(req.CountryCode, "62"),
		StateJson:   stringx.FirstNonEmpty(req.StateJSON, mapString(req.Data, "state_json")),
	})
	if err != nil {
		return nil, err
	}
	success := resp.GetAvailable() && resp.GetStatus() != "error" && resp.GetStatus() != "rate_limited"
	data := map[string]any{
		"available":          resp.GetAvailable(),
		"status":             resp.GetStatus(),
		"proxy_hash":         resp.GetProxyHash(),
		"device_fingerprint": resp.GetDeviceFingerprint(),
		"state_json":         resp.GetStateJson(),
	}
	out := h.baseResult(req, "gopay_app_check_phone", success, data)
	out.Phone = req.Phone
	out.StateJSON = stringx.FirstNonEmpty(resp.GetStateJson(), req.StateJSON, "{}")
	out.PhoneAccepted = success
	out.RotatableFailure = resp.GetStatus() == "registered"
	out.RetryableFailure = resp.GetStatus() == "error" || resp.GetStatus() == "rate_limited"
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) discardSignupPhone(req gopayActionRequest) *gopayActionResult {
	data := map[string]any{
		"activation_id": req.ActivationID,
		"reason":        strings.TrimSpace(req.Reason),
	}
	return h.baseResult(req, "gopay_app_signup_phone_cancel", true, data)
}

func (h gopayHTTPHandler) startSignup(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.SignupStart(ctx, &pb.SignupStartRequest{
		Phone:          req.Phone,
		CountryCode:    req.CountryCode,
		OtpChannel:     req.OTPChannel,
		SkipPhoneProbe: req.SkipPhoneProbe || anyBool(req.Data["skip_phone_probe"]),
		StateJson:      stringx.FirstNonEmpty(req.StateJSON, "{}"),
	})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "signup", resp.GetSuccess(), nil)
	out.applyOTP(resp.GetOtpSent(), resp.GetVerificationMethod())
	out.StateJSON = resp.GetStateJson()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
	return out, nil
}

func (h gopayHTTPHandler) retrySignup(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.SignupRetry(ctx, &pb.SignupRetryRequest{StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "signup", resp.GetSuccess(), nil)
	out.applyOTP(resp.GetOtpSent(), "")
	out.StateJSON = resp.GetStateJson()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
	return out, nil
}

func (h gopayHTTPHandler) completeSignup(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.SignupComplete(ctx, &pb.SignupCompleteRequest{Otp: req.otpValue(), StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "signup", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.Phone = resp.GetPhone()
	out.SignupComplete = resp.GetSuccess()
	out.RequirePIN = resp.GetPinSetupRequired()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfReady(ctx, req.withOTPChannel(out.OTPChannel), out)
	return out, nil
}
