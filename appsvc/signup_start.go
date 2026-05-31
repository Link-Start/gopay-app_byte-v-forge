package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (s *Server) startSignup(ctx context.Context, state stateMap, phone, name, email, countryCode, otpChannel string, skipPhoneProbe bool) map[string]any {
	cc := phoneCountryCode(s.cfg, countryCode)
	normalized := normalizePhoneWithConfig(s.cfg, phone, cc)
	if normalized == "" {
		return map[string]any{"success": false, "error": "signup phone missing"}
	}
	name, email = s.signupProfile(normalized, name, email)
	if name == "" {
		return map[string]any{"success": false, "error": "signup name missing"}
	}
	if cooldown := s.signupCooldownResult(state); cooldown != nil {
		return cooldown
	}
	s.clearSignupState(state, "")
	s.clearLoginState(state, "")
	device, err := s.ensureDevice(state)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if strings.TrimSpace(os.Getenv("GOPAY_APP_VERSION")) == "" && !strings.HasPrefix(strings.TrimSpace(device.AppVersion), "2.7.") {
		next, rawDevice, err := s.newLogonDevice()
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}
		device = next
		state["device"] = rawDevice
	}
	if probeTransactionID := stateString(state, "_signup_probe_transaction_id"); probeTransactionID != "" {
		device.TransactionID = probeTransactionID
	} else {
		state["_signup_probe_transaction_id"] = device.TransactionID
	}
	state["device"] = deviceToMap(device)
	deleteKeys(state, activeTokenKeys...)
	deleteKeys(state, activeTokenMetaKeys...)
	deleteKeys(state, tmpTokenKeys...)
	deleteKeys(state, tmpTokenMetaKeys...)
	state["_signup_phone"] = normalized
	state["_signup_country_code"] = cc
	state["_signup_name"] = name
	state["_signup_email"] = email
	state["_signup_started_at"] = time.Now().Unix()
	state["_signup_skip_phone_probe"] = skipPhoneProbe
	state["stage"] = "signup"
	delete(state, "last_error")
	if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{CountryCode: cc}); err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	proxyURL := s.proxyForState(state)
	client, err := s.newClient(ctx, "", proxyURL, device)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	supportWarmup := map[string]any{"attempted": true}
	if warmupResp, err := client.Customer.InitiateSupportCustomer(ctx); err != nil {
		supportWarmup["success"] = false
		supportWarmup["error"] = err.Error()
		state["_signup_support_warmup_error"] = err.Error()
	} else {
		status := 0
		if warmupResp != nil {
			status = warmupResp.StatusCode
		}
		supportWarmup["success"] = status >= 200 && status < 300
		supportWarmup["status_code"] = status
		state["_signup_support_warmup_status"] = status
		delete(state, "_signup_support_warmup_error")
	}
	state["_signup_support_warmup_at"] = time.Now().Unix()
	if skipPhoneProbe {
		state["_signup_phone_probe_skipped"] = true
		supportWarmup["phone_probe_skipped"] = true
	} else {
		probeResp, err := client.Auth.Post(ctx, "/goto-auth/login/methods", signupProbeBody{
			PhoneNumber:               normalized,
			CountryCode:               cc,
			Email:                     "",
			DeviceVerificationTokenID: "",
			ClientID:                  s.cfg.GotoClientID,
			ClientSecret:              s.cfg.GotoClientSecret,
		})
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}
		if probeResp.StatusCode == http.StatusOK || probeResp.StatusCode == http.StatusCreated {
			return map[string]any{"success": false, "error": "PHONE_REGISTERED", "raw_json": safeJSON(probeResp.Payload)}
		}
		if isRateLimited(probeResp) {
			return s.signupRateLimitResult(state, signupRateLimitScopeProbe, normalized, cc, rateLimitLabel(signupRateLimitScopeProbe), probeResp)
		}
		if !loginMethodsInvalidUser(probeResp) && probeResp.StatusCode >= http.StatusBadRequest {
			return map[string]any{"success": false, "error": apiError("signup phone probe failed", probeResp), "support_warmup": supportWarmup, "raw_json": safeJSON(probeResp.Payload)}
		}
	}
	device = device.WithNewTransactionID()
	state["_signup_cvs_transaction_id"] = device.TransactionID
	state["device"] = deviceToMap(device)
	client, err = s.newClient(ctx, "", proxyURL, device)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error(), "support_warmup": supportWarmup}
	}
	methodsResp, err := client.Auth.Post(ctx, "/cvs/v1/methods", signupMethodsBody{
		CountryCode:               cc,
		DeviceVerificationTokenID: nil,
		EmailAddress:              nil,
		Flow:                      "signup",
		PhoneNumber:               normalized,
		ClientID:                  s.cfg.GotoClientID,
		ClientSecret:              s.cfg.GotoClientSecret,
	})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if isRateLimited(methodsResp) {
		return s.signupRateLimitResult(state, signupRateLimitScopeMethods, normalized, cc, rateLimitLabel(signupRateLimitScopeMethods), methodsResp)
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
	method := chooseOTPMethod(methods, otpChannel, stringx.FirstNonEmpty(defaultMethod, "otp_wa"))
	if method == "" {
		return map[string]any{"success": false, "error": fmt.Sprintf("otp method unavailable: %v", methods), "response_shape": responseShape(methodsResp)}
	}
	initiateDelay := s.signupInitiateDelay()
	if initiateDelay > 0 {
		now := time.Now().Unix()
		state["_signup_initiate_delay_seconds"] = int64(initiateDelay.Seconds())
		state["_signup_initiate_delay_started_at"] = now
		if err := sleepWithContext(ctx, initiateDelay); err != nil {
			return map[string]any{"success": false, "error": err.Error(), "signup_initiate_delay_seconds": int64(initiateDelay.Seconds())}
		}
		state["_signup_initiate_delay_finished_at"] = time.Now().Unix()
	}
	initResp, err := client.Auth.Request(ctx, http.MethodPost, "/cvs/v1/initiate", signupInitiateBody{
		CountryCode:               cc,
		DeviceVerificationTokenID: nil,
		EmailAddress:              nil,
		Flow:                      "signup",
		IsMultipleMethod:          nil,
		PhoneNumber:               normalized,
		VerificationID:            verificationID,
		VerificationMethod:        method,
		ClientID:                  s.cfg.GotoClientID,
		ClientSecret:              s.cfg.GotoClientSecret,
	}, nil)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if isRateLimited(initResp) {
		return s.signupRateLimitResult(state, signupRateLimitScopeInitiate, normalized, cc, rateLimitLabel(signupRateLimitScopeInitiate), initResp)
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
