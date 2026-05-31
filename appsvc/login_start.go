package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/common-lib/stringx"
	gopayapp "github.com/byte-v-forge/gopay-app/protocol/app"
)

func (s *Server) startLogin(ctx context.Context, state stateMap, phone, pin, countryCode, otpChannel string) map[string]any {
	cc := phoneCountryCode(s.cfg, countryCode)
	normalized := normalizePhoneWithConfig(s.cfg, phone, cc)
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
				time.Sleep(loginMethodsBackoff(attempt))
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
			time.Sleep(loginMethodsBackoff(attempt))
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
	if !contains(methods, "goto_pin") {
		return map[string]any{"success": false, "error": fmt.Sprintf("goto_pin unavailable: %v", methods)}
	}
	if strings.TrimSpace(pin) == "" {
		return map[string]any{"success": false, "error": "gopay pin missing"}
	}
	c := client
	initResp, err := c.Auth.Request(ctx, http.MethodPost, "/cvs/v1/initiate", signupInitiateBody{
		VerificationID:            verificationID,
		Flow:                      "login_1fa",
		VerificationMethod:        "goto_pin",
		CountryCode:               cc,
		EmailAddress:              nil,
		ClientID:                  s.cfg.GotoClientID,
		PhoneNumber:               normalized,
		ClientSecret:              s.cfg.GotoClientSecret,
		IsMultipleMethod:          true,
		DeviceVerificationTokenID: nil,
	}, http.Header{"Authorization": []string{""}})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if initResp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("login pin initiate failed", initResp)}
	}
	challengeID := challengeIDFrom(initResp.Data())
	if challengeID == "" {
		shape := responseShape(initResp)
		return map[string]any{"success": false, "error": "pin challenge_id missing: " + safeJSON(shape), "response_shape": shape}
	}
	if pinPage, err := c.Customer.Get(ctx, "/api/v2/challenges/"+challengeID+"/pin-page/nb"); err != nil || pinPage.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("pin page failed", pinPage)}
	}
	pinResp, err := c.Customer.Post(ctx, "/api/v1/users/pin/tokens/nb", map[string]any{
		"challenge_id": challengeID,
		"client_id":    s.cfg.PINClientID,
		"pin":          pin,
	})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if pinResp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("pin token failed", pinResp)}
	}
	validationJWT := stringForAnyKey(pinResp.Data(), "token")
	if validationJWT == "" {
		return map[string]any{"success": false, "error": "pin validation token missing"}
	}
	verifyResp, err := c.Auth.Post(ctx, "/cvs/v1/verify", cvsVerifyBody{
		Data:               map[string]any{"challenge_id": challengeID, "validation_jwt": validationJWT},
		Flow:               "login_1fa",
		VerificationID:     verificationID,
		VerificationMethod: "goto_pin",
		ClientID:           s.cfg.GotoClientID,
		ClientSecret:       s.cfg.GotoClientSecret,
	})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if verifyResp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("login pin verify failed", verifyResp)}
	}
	verificationToken := verificationTokenFrom(verifyResp.Data())
	if verificationToken == "" {
		return map[string]any{"success": false, "error": "1fa verification_token missing"}
	}
	accountResp, err := c.Auth.Request(ctx, http.MethodPost, "/goto-auth/accountlist", gotoAuthClientBody{
		ClientID:     s.cfg.GotoClientID,
		ClientSecret: s.cfg.GotoClientSecret,
	}, http.Header{"Verification-Token": []string{"Bearer " + verificationToken}})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if accountResp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("accountlist failed", accountResp)}
	}
	accountID := firstAccountID(accountListFrom(accountResp.Data()))
	oneFAToken := oneFATokenFrom(accountResp.Data())
	if accountID == "" || oneFAToken == "" {
		return map[string]any{"success": false, "error": "account_id or 1fa_token missing"}
	}
	tokenResp, err := c.Auth.Post(ctx, "/goto-auth/token", gotoCVSTokenBody{
		AccountID:    accountID,
		ExtUserToken: nil,
		GrantType:    "cvs",
		Token:        oneFAToken,
		ClientID:     s.cfg.GotoClientID,
		ClientSecret: s.cfg.GotoClientSecret,
	})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if tokenResp.StatusCode == http.StatusCreated {
		s.persistLoginReady(state, tokenResp.Data(), normalized)
		return map[string]any{"success": true, "ready": true, "otp_sent": false}
	}
	twoFAToken := twoFATokenFrom(tokenResp.Data())
	verificationID = verificationIDFrom(tokenResp.Data())
	if tokenResp.StatusCode != http.StatusForbidden || twoFAToken == "" || verificationID == "" {
		return map[string]any{"success": false, "error": apiError("token exchange failed", tokenResp)}
	}
	otpMethods := methodsFrom(tokenResp.Data())
	method := chooseOTPMethod(otpMethods, otpChannel, "otp_wa")
	if method == "" {
		return map[string]any{"success": false, "error": fmt.Sprintf("otp method unavailable: %v", otpMethods), "response_shape": responseShape(tokenResp)}
	}
	otpResp, err := c.Auth.Request(ctx, http.MethodPost, "/cvs/v1/initiate", signupInitiateBody{
		VerificationID:            verificationID,
		Flow:                      "login_2fa",
		VerificationMethod:        method,
		CountryCode:               cc,
		EmailAddress:              nil,
		ClientID:                  s.cfg.GotoClientID,
		PhoneNumber:               normalized,
		ClientSecret:              s.cfg.GotoClientSecret,
		IsMultipleMethod:          nil,
		DeviceVerificationTokenID: nil,
	}, http.Header{"Authorization": []string{""}})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if otpResp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("2fa otp initiate failed", otpResp)}
	}
	otpToken := otpTokenFrom(otpResp.Data())
	if otpToken == "" {
		return map[string]any{"success": false, "error": "2fa otp_token missing"}
	}
	s.persistLoginOTP(state, normalized, cc, verificationID, method, otpToken, twoFAToken, "login_2fa")
	return map[string]any{"success": true, "ready": false, "otp_sent": true, "verification_id": verificationID, "method": method}
}
