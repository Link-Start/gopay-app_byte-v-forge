package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/common-lib/timex"
	gopayapp "github.com/byte-v-forge/gopay-app/protocol/app"
)

func (s *Server) requestSignupOTP(ctx context.Context, state stateMap, device gopayapp.DeviceFingerprint, proxyURL string, input signupStartInput, supportWarmup map[string]any) map[string]any {
	device = device.WithNewTransactionID()
	state["_signup_cvs_transaction_id"] = device.TransactionID
	state["device"] = deviceToMap(device)
	client, err := s.newClient(ctx, "", proxyURL, device)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error(), "support_warmup": supportWarmup}
	}
	methodsResp, err := client.Auth.Post(ctx, "/cvs/v1/methods", signupMethodsBody{
		CountryCode:               input.CountryCode,
		DeviceVerificationTokenID: nil,
		EmailAddress:              nil,
		Flow:                      "signup",
		PhoneNumber:               input.Phone,
		ClientID:                  s.cfg.GotoClientID,
		ClientSecret:              s.cfg.GotoClientSecret,
	})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if isRateLimited(methodsResp) {
		return s.signupRateLimitResult(state, signupRateLimitScopeMethods, input.Phone, input.CountryCode, rateLimitLabel(signupRateLimitScopeMethods), methodsResp)
	}
	if methodsResp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("signup methods failed", methodsResp), "raw_json": safeJSON(methodsResp.Payload)}
	}
	methodsData := methodsResp.Data()
	verificationID := verificationIDFrom(methodsData)
	if verificationID == "" {
		shape := responseShape(methodsResp)
		return map[string]any{"success": false, "error": "signup verification_id missing: " + safeJSON(shape), "response_shape": shape}
	}
	methods := methodsFrom(methodsData)
	defaultMethod := stringForAnyKey(methodsData, "default_method", "defaultMethod")
	method := chooseOTPMethod(methods, input.OTPChannel, stringx.FirstNonEmpty(defaultMethod, "otp_wa"))
	if method == "" {
		return map[string]any{"success": false, "error": fmt.Sprintf("otp method unavailable: %v", methods), "response_shape": responseShape(methodsResp)}
	}
	initiateDelay := s.signupInitiateDelay()
	if result := sleepBeforeSignupInitiate(ctx, state, initiateDelay); result != nil {
		return result
	}
	initResp, err := client.Auth.Request(ctx, http.MethodPost, "/cvs/v1/initiate", signupInitiateBody{
		CountryCode:               input.CountryCode,
		DeviceVerificationTokenID: nil,
		EmailAddress:              nil,
		Flow:                      "signup",
		IsMultipleMethod:          nil,
		PhoneNumber:               input.Phone,
		VerificationID:            verificationID,
		VerificationMethod:        method,
		ClientID:                  s.cfg.GotoClientID,
		ClientSecret:              s.cfg.GotoClientSecret,
	}, nil)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if isRateLimited(initResp) {
		return s.signupRateLimitResult(state, signupRateLimitScopeInitiate, input.Phone, input.CountryCode, rateLimitLabel(signupRateLimitScopeInitiate), initResp)
	}
	if initResp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("signup otp initiate failed", initResp), "method": method, "raw_json": safeJSON(initResp.Payload)}
	}
	otpToken := otpTokenFrom(initResp.Data())
	if otpToken == "" {
		return map[string]any{"success": false, "error": "signup otp_token missing", "raw_json": safeJSON(initResp.Payload)}
	}
	s.persistSignupOTP(state, verificationID, method, otpToken)
	return map[string]any{
		"success": true, "otp_sent": true, "verification_id": verificationID,
		"method": method, "default_method": defaultMethod, "retry_timer_seconds": initResp.Data()["retry_timer_in_seconds"],
		"signup_initiate_delay_seconds": int64(initiateDelay.Seconds()),
		"support_warmup":                supportWarmup,
		"raw_json":                      safeJSON(initResp.Payload),
	}
}

func sleepBeforeSignupInitiate(ctx context.Context, state stateMap, initiateDelay time.Duration) map[string]any {
	if initiateDelay <= 0 {
		return nil
	}
	now := time.Now().Unix()
	state["_signup_initiate_delay_seconds"] = int64(initiateDelay.Seconds())
	state["_signup_initiate_delay_started_at"] = now
	if err := timex.Sleep(ctx, initiateDelay); err != nil {
		return map[string]any{"success": false, "error": err.Error(), "signup_initiate_delay_seconds": int64(initiateDelay.Seconds())}
	}
	state["_signup_initiate_delay_finished_at"] = time.Now().Unix()
	return nil
}
