package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	gopayapp "github.com/byte-v-forge/gopay-app/protocol/app"
)

type loginPINStart struct {
	Client         *gopayapp.ClientSet
	Phone          string
	CountryCode    string
	VerificationID string
	Methods        []string
	PIN            string
	OTPChannel     string
}

func (s *Server) startLoginWithPIN(ctx context.Context, state stateMap, input loginPINStart) map[string]any {
	if !contains(input.Methods, "goto_pin") {
		return map[string]any{"success": false, "error": fmt.Sprintf("goto_pin unavailable: %v", input.Methods)}
	}
	if strings.TrimSpace(input.PIN) == "" {
		return map[string]any{"success": false, "error": "gopay pin missing"}
	}
	c := input.Client
	initResp, err := c.Auth.Request(ctx, http.MethodPost, "/cvs/v1/initiate", signupInitiateBody{
		VerificationID:            input.VerificationID,
		Flow:                      "login_1fa",
		VerificationMethod:        "goto_pin",
		CountryCode:               input.CountryCode,
		EmailAddress:              nil,
		ClientID:                  s.cfg.GotoClientID,
		PhoneNumber:               input.Phone,
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
		"pin":          input.PIN,
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
		VerificationID:     input.VerificationID,
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
		s.persistLoginReady(state, tokenResp.Data(), input.Phone)
		return map[string]any{"success": true, "ready": true, "otp_sent": false}
	}
	twoFAToken := twoFATokenFrom(tokenResp.Data())
	verificationID := verificationIDFrom(tokenResp.Data())
	if tokenResp.StatusCode != http.StatusForbidden || twoFAToken == "" || verificationID == "" {
		return map[string]any{"success": false, "error": apiError("token exchange failed", tokenResp)}
	}
	otpMethods := methodsFrom(tokenResp.Data())
	method := chooseOTPMethod(otpMethods, input.OTPChannel, "otp_wa")
	if method == "" {
		return map[string]any{"success": false, "error": fmt.Sprintf("otp method unavailable: %v", otpMethods), "response_shape": responseShape(tokenResp)}
	}
	otpResp, err := c.Auth.Request(ctx, http.MethodPost, "/cvs/v1/initiate", signupInitiateBody{
		VerificationID:            verificationID,
		Flow:                      "login_2fa",
		VerificationMethod:        method,
		CountryCode:               input.CountryCode,
		EmailAddress:              nil,
		ClientID:                  s.cfg.GotoClientID,
		PhoneNumber:               input.Phone,
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
	s.persistLoginOTP(state, input.Phone, input.CountryCode, verificationID, method, otpToken, twoFAToken, "login_2fa")
	return map[string]any{"success": true, "ready": false, "otp_sent": true, "verification_id": verificationID, "method": method}
}
