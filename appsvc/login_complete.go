package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/common-lib/stringx"
	gopayapp "github.com/byte-v-forge/gopay-app/protocol/app"
)

func (s *Server) completeLogin(ctx context.Context, state stateMap, otp string) error {
	if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{}); err != nil {
		return err
	}
	device, err := s.ensureDevice(state)
	if err != nil {
		return err
	}
	client, err := s.newClient(ctx, "", s.proxyForState(state), device)
	if err != nil {
		return err
	}
	verificationID := stateString(state, "_login_verification_id")
	otpToken := stateString(state, "_login_otp_token")
	method := stringx.FirstNonEmpty(stateString(state, "_login_verification_method"), "otp_wa")
	flow := stringx.FirstNonEmpty(stateString(state, "_login_flow"), "login_2fa")
	twoFAToken := stateString(state, "_login_2fa_token")
	if verificationID == "" || otpToken == "" {
		return fmt.Errorf("login otp state missing")
	}
	if flow == "login_2fa" && twoFAToken == "" {
		return fmt.Errorf("login 2fa state missing")
	}
	verifyResp, err := client.Auth.Post(ctx, "/cvs/v1/verify", cvsVerifyBody{
		Data:               map[string]any{"otp": strings.TrimSpace(otp), "otp_token": otpToken},
		Flow:               flow,
		VerificationID:     verificationID,
		VerificationMethod: method,
		ClientID:           s.cfg.GotoClientID,
		ClientSecret:       s.cfg.GotoClientSecret,
	})
	if err != nil {
		return err
	}
	if verifyResp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", apiError(flow+" verify failed", verifyResp))
	}
	verificationToken := verificationTokenFrom(verifyResp.Data())
	if verificationToken == "" {
		return fmt.Errorf("%s verification_token missing", flow)
	}
	if flow == "login_1fa" {
		accountResp, err := client.Auth.Request(ctx, http.MethodPost, "/goto-auth/accountlist", gotoAuthClientBody{
			ClientID:     s.cfg.GotoClientID,
			ClientSecret: s.cfg.GotoClientSecret,
		}, http.Header{"Verification-Token": []string{"Bearer " + verificationToken}})
		if err != nil {
			return err
		}
		if accountResp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s", apiError("accountlist failed", accountResp))
		}
		accountID := firstAccountID(accountListFrom(accountResp.Data()))
		oneFAToken := oneFATokenFrom(accountResp.Data())
		if accountID == "" || oneFAToken == "" {
			return fmt.Errorf("account_id or 1fa_token missing")
		}
		tokenResp, err := client.Auth.Post(ctx, "/goto-auth/token", gotoCVSTokenBody{
			AccountID:    accountID,
			ExtUserToken: nil,
			GrantType:    "cvs",
			Token:        oneFAToken,
			ClientID:     s.cfg.GotoClientID,
			ClientSecret: s.cfg.GotoClientSecret,
		})
		if err != nil {
			return err
		}
		if tokenResp.StatusCode != http.StatusCreated {
			if tokenResp.StatusCode == http.StatusForbidden && twoFATokenFrom(tokenResp.Data()) != "" {
				return s.continueLogin2FA(ctx, client, state, tokenResp)
			}
			return fmt.Errorf("%s", apiError("cvs token failed", tokenResp))
		}
		s.persistLoginReady(state, tokenResp.Data(), stateString(state, "_login_phone"))
		return nil
	}
	tokenResp, err := client.Auth.Request(ctx, http.MethodPost, "/goto-auth/token", gotoChallengeTokenBody{
		ExtUserToken: nil,
		GrantType:    "challenge",
		Token:        twoFAToken,
		ClientID:     s.cfg.GotoClientID,
		ClientSecret: s.cfg.GotoClientSecret,
	}, http.Header{"Verification-Token": []string{"Bearer " + verificationToken}})
	if err != nil {
		return err
	}
	if tokenResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("%s", apiError("challenge token failed", tokenResp))
	}
	s.persistLoginReady(state, tokenResp.Data(), stateString(state, "_login_phone"))
	return nil
}

func (s *Server) continueLogin2FA(ctx context.Context, client *gopayapp.ClientSet, state stateMap, tokenResp *httpjson.Response) error {
	twoFAToken := twoFATokenFrom(tokenResp.Data())
	verificationID := verificationIDFrom(tokenResp.Data())
	if twoFAToken == "" || verificationID == "" {
		return fmt.Errorf("%s", apiError("cvs token 2fa challenge missing", tokenResp))
	}
	otpMethods := methodsFrom(tokenResp.Data())
	defaultMethod := stringForAnyKey(tokenResp.Data(), "default_method", "defaultMethod")
	previousMethod := stateString(state, "_login_verification_method")
	method := chooseOTPMethod(otpMethods, "", stringx.FirstNonEmpty(defaultMethod, previousMethod, "otp_wa"))
	if method == "" {
		return fmt.Errorf("2fa otp method unavailable: %v", otpMethods)
	}
	phone := stateString(state, "_login_phone")
	countryCode := stateString(state, "_login_country_code")
	otpResp, err := client.Auth.Request(ctx, http.MethodPost, "/cvs/v1/initiate", signupInitiateBody{
		VerificationID:            verificationID,
		Flow:                      "login_2fa",
		VerificationMethod:        method,
		CountryCode:               countryCode,
		EmailAddress:              nil,
		ClientID:                  s.cfg.GotoClientID,
		PhoneNumber:               phone,
		ClientSecret:              s.cfg.GotoClientSecret,
		IsMultipleMethod:          nil,
		DeviceVerificationTokenID: nil,
	}, http.Header{"Authorization": []string{""}})
	if err != nil {
		return err
	}
	if otpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", apiError("2fa otp initiate failed", otpResp))
	}
	otpToken := otpTokenFrom(otpResp.Data())
	if otpToken == "" {
		return fmt.Errorf("2fa otp_token missing")
	}
	s.persistLoginOTP(state, phone, countryCode, verificationID, method, otpToken, twoFAToken, "login_2fa")
	return nil
}
