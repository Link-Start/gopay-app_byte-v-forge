package appsvc

import (
	"context"
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/common-lib/timex"
	gopayapp "github.com/byte-v-forge/gopay-app/protocol/app"
)

func (s *Server) startLogin(ctx context.Context, state stateMap, phone, pin, countryCode, otpChannel string) map[string]any {
	cc := phoneCountryCode(s.cfg, countryCode)
	normalized := normalizePhoneWithConfig(s.cfg, phone, cc)
	pin = s.resolveGoPayAccountPin(ctx, state, pin)
	attempts := loginMethodsMaxAttempts
	var resp *httpjson.Response
	var client *gopayapp.ClientSet
	var methods []string
	var defaultMethod string
	var verificationID string
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			if err := s.rotateLoginAttemptIdentity(ctx, state); err != nil {
				return map[string]any{"success": false, "error": err.Error()}
			}
		} else if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{CountryCode: cc}); err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}
		proxyURL := s.proxyForState(state)
		device, err := s.ensureDevice(state)
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}
		state["_login_phone"] = normalized
		state["_login_country_code"] = cc
		state["_login_started_at"] = time.Now().Unix()
		state["stage"] = "login"
		delete(state, "last_error")
		c, err := s.newClient(ctx, "", proxyURL, device)
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}
		if probeID, probeMethods, probeDefault, ok := s.reusableLoginProbe(state, normalized, cc); ok {
			client = c
			verificationID = probeID
			methods = probeMethods
			defaultMethod = probeDefault
			break
		}
		resp, err = c.Auth.Post(ctx, "/goto-auth/login/methods", signupProbeBody{
			PhoneNumber:               normalized,
			CountryCode:               cc,
			Email:                     "",
			DeviceVerificationTokenID: "",
			ClientID:                  s.cfg.GotoClientID,
			ClientSecret:              s.cfg.GotoClientSecret,
		})
		if err != nil {
			if attempt < attempts && retryableGoPayTransportError(err) {
				if sleepErr := timex.Sleep(ctx, loginMethodsBackoff(attempt)); sleepErr != nil {
					return map[string]any{"success": false, "error": sleepErr.Error()}
				}
				continue
			}
			return map[string]any{"success": false, "error": err.Error()}
		}
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			client = c
			verificationID, methods, defaultMethod = s.persistLoginProbe(state, normalized, cc, resp.Data())
			break
		}
		if isRateLimited(resp) && attempt < attempts {
			if sleepErr := timex.Sleep(ctx, loginMethodsBackoff(attempt)); sleepErr != nil {
				return map[string]any{"success": false, "error": sleepErr.Error()}
			}
			continue
		}
		if isRateLimited(resp) {
			return map[string]any{"success": false, "error": loginMethodsRateLimitedError()}
		}
		if loginMethodsInvalidUser(resp) {
			return map[string]any{"success": false, "not_registered": true, "error": apiError("login methods failed", resp)}
		}
		return map[string]any{"success": false, "error": apiError("login methods failed", resp)}
	}
	if client == nil {
		return map[string]any{"success": false, "error": "login methods failed"}
	}
	if verificationID == "" {
		if resp != nil {
			shape := responseShape(resp)
			return map[string]any{"success": false, "error": "verification_id missing: " + safeJSON(shape), "response_shape": shape}
		}
		return map[string]any{"success": false, "error": "verification_id missing from login probe state"}
	}
	if method := chooseOTPMethod(methods, otpChannel, stringx.FirstNonEmpty(defaultMethod, "otp_wa")); method != "" {
		otpResp, err := client.Auth.Request(ctx, http.MethodPost, "/cvs/v1/initiate", signupInitiateBody{
			VerificationID:            verificationID,
			Flow:                      "login_1fa",
			VerificationMethod:        method,
			CountryCode:               cc,
			EmailAddress:              nil,
			ClientID:                  s.cfg.GotoClientID,
			PhoneNumber:               normalized,
			ClientSecret:              s.cfg.GotoClientSecret,
			IsMultipleMethod:          nil,
			DeviceVerificationTokenID: nil,
		}, nil)
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}
		if otpResp.StatusCode != http.StatusOK {
			return map[string]any{"success": false, "error": apiError("login otp initiate failed", otpResp)}
		}
		otpToken := otpTokenFrom(otpResp.Data())
		if otpToken == "" {
			return map[string]any{"success": false, "error": "login otp_token missing", "response_shape": responseShape(otpResp)}
		}
		s.persistLoginOTP(state, normalized, cc, verificationID, method, otpToken, "", "login_1fa")
		return map[string]any{"success": true, "ready": false, "otp_sent": true, "verification_id": verificationID, "method": method}
	}
	return s.startLoginWithPIN(ctx, state, loginPINStart{
		Client:         client,
		Phone:          normalized,
		CountryCode:    cc,
		VerificationID: verificationID,
		Methods:        methods,
		PIN:            pin,
		OTPChannel:     otpChannel,
	})
}
