package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (s *Server) retrySignupOTP(ctx context.Context, state stateMap) map[string]any {
	if stateString(state, "stage") != "signup_otp_pending" {
		return map[string]any{"success": false, "error": fmt.Sprintf("not waiting for signup otp: %s", stringx.FirstNonEmpty(stateString(state, "stage"), "idle"))}
	}
	otpToken := stateString(state, "_signup_otp_token")
	method := stringx.FirstNonEmpty(stateString(state, "_signup_verification_method"), "otp_sms")
	if otpToken == "" {
		return map[string]any{"success": false, "error": "signup otp state missing"}
	}
	device, err := s.ensureDevice(state)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{}); err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	client, err := s.newClient(ctx, "", s.proxyForState(state), device)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	resp, err := client.Auth.Post(ctx, "/cvs/v1/retry", s.authBody(map[string]any{
		"flow":                "signup",
		"verification_method": method,
		"data":                map[string]any{"otp_token": otpToken},
	}))
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if resp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("signup otp retry failed", resp), "raw_json": safeJSON(resp.Payload)}
	}
	if newToken := otpTokenFrom(resp.Data()); newToken != "" {
		state["_signup_otp_token"] = newToken
	}
	now := time.Now().Unix()
	if channel := normalizeActionOTPChannel(method); channel != "" {
		state["_otp_channel"] = channel
	}
	state["_signup_otp_sent_at"] = now
	state["_signup_otp_expires_at"] = now + int64(s.cfg.OTPTimeout.Seconds())
	state["stage"] = "signup_otp_pending"
	delete(state, "last_error")
	return map[string]any{"success": true, "otp_sent": true, "raw_json": safeJSON(resp.Payload)}
}

func (s *Server) completeSignup(ctx context.Context, state stateMap, otp string) map[string]any {
	if stateString(state, "stage") != "signup_otp_pending" {
		return map[string]any{"success": false, "error": fmt.Sprintf("not waiting for signup otp: %s", stringx.FirstNonEmpty(stateString(state, "stage"), "idle"))}
	}
	otp = strings.TrimSpace(otp)
	if otp == "" {
		return map[string]any{"success": false, "error": "signup otp required"}
	}
	phone := stateString(state, "_signup_phone")
	cc := stringx.FirstNonEmpty(stateString(state, "_signup_country_code"), phoneCountryCode(s.cfg, ""))
	name := stateString(state, "_signup_name")
	email := stateString(state, "_signup_email")
	verificationID := stateString(state, "_signup_verification_id")
	method := stringx.FirstNonEmpty(stateString(state, "_signup_verification_method"), "otp_sms")
	otpToken := stateString(state, "_signup_otp_token")
	if phone == "" || verificationID == "" || otpToken == "" {
		return map[string]any{"success": false, "error": "signup otp state missing"}
	}
	device, err := s.ensureDevice(state)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{}); err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	client, err := s.newClient(ctx, "", s.proxyForState(state), device)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	verifyResp, err := client.Auth.Post(ctx, "/cvs/v1/verify", s.authBody(map[string]any{
		"data":                map[string]any{"otp": otp, "otp_token": otpToken},
		"flow":                "signup",
		"verification_id":     verificationID,
		"verification_method": method,
	}))
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if verifyResp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("signup otp verify failed", verifyResp), "raw_json": safeJSON(verifyResp.Payload)}
	}
	verificationToken := verificationTokenFrom(verifyResp.Data())
	if verificationToken == "" {
		return map[string]any{"success": false, "error": "signup verification_token missing", "raw_json": safeJSON(verifyResp.Payload)}
	}
	signupResp, err := client.Gojek.Request(ctx, http.MethodPost, "/v7/customers/signup", map[string]any{
		"client_name":   s.cfg.GotoClientID,
		"client_secret": s.cfg.GotoClientSecret,
		"data": map[string]any{
			"name":               name,
			"phone":              cc + phone,
			"email":              email,
			"signed_up_country":  cc,
			"onboarding_partner": "gopay_consumer_app",
		},
	}, http.Header{
		"Authorization":      []string{s.signupBasicAuthorization()},
		"Verification-Token": []string{"Bearer " + verificationToken},
	})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if signupResp.StatusCode != http.StatusCreated {
		return map[string]any{"success": false, "error": apiError("customer signup failed", signupResp), "raw_json": safeJSON(signupResp.Payload)}
	}
	s.storeTokenResponse(state, signupResp.Data(), false)
	if stateString(state, "token") == "" {
		state["last_error"] = "signup access token missing"
		return map[string]any{"success": false, "error": stateString(state, "last_error"), "raw_json": safeJSON(signupResp.Payload)}
	}
	state["phone"] = phone
	state["name"] = name
	state["email"] = email
	state["stage"] = "signup_pin_required"
	delete(state, "last_error")
	deleteKeys(state, signupOTPStateKeys...)
	return map[string]any{"success": true, "phone": phone, "pin_setup_required": true, "raw_json": safeJSON(signupResp.Payload)}
}
