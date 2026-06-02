package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/common-lib/timex"
)

func (s *Server) startSignupPIN(ctx context.Context, state stateMap, pin, otpChannel string) map[string]any {
	var last map[string]any
	for attempt := 1; attempt <= 2; attempt++ {
		if attempt > 1 {
			if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{ForceNew: true}); err != nil {
				return map[string]any{"success": false, "error": err.Error()}
			}
		}
		last = s.startSignupPINOnce(ctx, state, pin, otpChannel)
		if anyBool(last["success"]) || !retryableGoPayActionError(last) {
			return last
		}
		if err := timex.Sleep(ctx, loginMethodsBackoff(attempt)); err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}
	}
	return last
}

func (s *Server) startSignupPINOnce(ctx context.Context, state stateMap, pin, otpChannel string) map[string]any {
	pin = s.resolveGoPayAccountPin(ctx, state, pin)
	if pin == "" {
		return map[string]any{"success": false, "error": "gopay pin missing"}
	}
	refresh := s.ensureAccessToken(ctx, state, s.cfg.TokenRefreshMinTTL, false)
	if !anyBool(refresh["success"]) && !tokenUsable(state, "token", 0) {
		return map[string]any{"success": false, "error": stringx.FirstNonEmpty(anyString(refresh["error"]), "token refresh failed")}
	}
	phone := stringx.FirstNonEmpty(stateString(state, "_signup_phone"), stateString(state, "phone"))
	if phone == "" {
		return map[string]any{"success": false, "error": "signup phone missing"}
	}
	client, err := s.newClientWithState(ctx, state, false)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	profile, _ := client.Customer.Get(ctx, "/v1/users/profile")
	pinSetup := false
	if profile != nil && profile.StatusCode == http.StatusOK {
		pinSetup, _ = pinSetupFlagFromProfileData(profile.Data())
	}
	if pinSetup {
		phone = stringx.FirstNonEmpty(stringForAnyKey(profile.Data(), "phone", "number"), phone)
		state["phone"] = normalizePhone(phone, "")
		state["stage"] = "ready"
		updatePINSetupState(state, true)
		state["ready_at"] = time.Now().Unix()
		delete(state, "last_error")
		deleteKeys(state, signupAccountStateKeys...)
		deleteKeys(state, signupOTPStateKeys...)
		deleteKeys(state, signupPINStateKeys...)
		return map[string]any{"success": true, "phone": stateString(state, "phone"), "pin_setup_complete": true}
	}
	allowed, err := client.Customer.Post(ctx, "/api/v1/users/pins/allowed", map[string]any{"pin": pin})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if allowed.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("pin allowed failed", allowed)}
	}
	methods, err := client.Auth.Post(ctx, "/cvs/v1/methods", s.authBody(map[string]any{
		"country_code":                 nil,
		"device_verification_token_id": nil,
		"email_address":                nil,
		"flow":                         "goto_pin_wa_sms",
		"phone_number":                 nil,
	}))
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if methods.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("pin otp methods failed", methods)}
	}
	methodsData := methods.Data()
	verificationID := verificationIDFrom(methodsData)
	if verificationID == "" {
		shape := responseShape(methods)
		return map[string]any{"success": false, "error": "pin verification_id missing: " + safeJSON(shape), "response_shape": shape}
	}
	otpMethods := methodsFrom(methodsData)
	method := chooseOTPMethod(otpMethods, otpChannel, "otp_sms")
	if method == "" {
		return map[string]any{"success": false, "error": fmt.Sprintf("otp method unavailable: %v", otpMethods), "response_shape": responseShape(methods)}
	}
	initResp, err := client.Auth.Post(ctx, "/cvs/v1/initiate", s.authBody(map[string]any{
		"country_code":                 nil,
		"device_verification_token_id": nil,
		"email_address":                nil,
		"flow":                         "goto_pin_wa_sms",
		"is_multiple_method":           nil,
		"phone_number":                 nil,
		"verification_id":              verificationID,
		"verification_method":          method,
	}))
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if initResp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("pin otp initiate failed", initResp)}
	}
	otpToken := otpTokenFrom(initResp.Data())
	if otpToken == "" {
		return map[string]any{"success": false, "error": "pin otp_token missing"}
	}
	now := time.Now().Unix()
	state["_signup_pin_challenge_id"] = ""
	state["_signup_pin_client_id"] = ""
	state["_signup_pin_verification_id"] = verificationID
	state["_signup_pin_verification_method"] = method
	if channel := normalizeActionOTPChannel(method); channel != "" {
		state["_otp_channel"] = channel
	}
	state["_signup_pin_otp_token"] = otpToken
	state["_signup_pin_otp_sent_at"] = now
	state["_signup_pin_otp_expires_at"] = now + int64(s.cfg.OTPTimeout.Seconds())
	state["stage"] = "signup_pin_otp_pending"
	delete(state, "last_error")
	return map[string]any{"success": true, "otp_sent": true, "verification_id": verificationID, "method": method}
}
