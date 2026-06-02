package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
)

func retryableGoPayActionError(result map[string]any) bool {
	err := strings.TrimSpace(anyString(result["error"]))
	if err == "" {
		return false
	}
	return retryableGoPayTransportError(fmt.Errorf("%s", err))
}

func (s *Server) retrySignupPIN(ctx context.Context, state stateMap) map[string]any {
	if stateString(state, "stage") != "signup_pin_otp_pending" {
		return map[string]any{"success": false, "error": fmt.Sprintf("not waiting for signup pin otp: %s", stringx.FirstNonEmpty(stateString(state, "stage"), "idle"))}
	}
	otpToken := stateString(state, "_signup_pin_otp_token")
	method := stringx.FirstNonEmpty(stateString(state, "_signup_pin_verification_method"), "otp_sms")
	if otpToken == "" {
		return map[string]any{"success": false, "error": "signup pin otp state missing"}
	}
	client, err := s.newClientWithState(ctx, state, true)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	resp, err := client.Auth.Post(ctx, "/cvs/v1/retry", s.authBody(map[string]any{
		"flow":                "goto_pin_wa_sms",
		"verification_method": method,
		"data":                map[string]any{"otp_token": otpToken},
	}))
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if resp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("pin otp retry failed", resp)}
	}
	newToken := otpTokenFrom(resp.Data())
	if newToken == "" {
		return map[string]any{"success": false, "error": "pin retry otp_token missing"}
	}
	now := time.Now().Unix()
	state["_signup_pin_otp_token"] = newToken
	if channel := normalizeActionOTPChannel(method); channel != "" {
		state["_otp_channel"] = channel
	}
	state["_signup_pin_otp_sent_at"] = now
	state["_signup_pin_otp_expires_at"] = now + int64(s.cfg.OTPTimeout.Seconds())
	delete(state, "last_error")
	return map[string]any{"success": true, "otp_sent": true}
}
