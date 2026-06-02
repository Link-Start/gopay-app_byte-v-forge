package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (s *Server) completeSignupPIN(ctx context.Context, state stateMap, otp, pin string) map[string]any {
	if stateString(state, "stage") != "signup_pin_otp_pending" {
		return map[string]any{"success": false, "error": fmt.Sprintf("not waiting for signup pin otp: %s", stringx.FirstNonEmpty(stateString(state, "stage"), "idle"))}
	}
	otp = strings.TrimSpace(otp)
	pin = s.resolveGoPayAccountPin(ctx, state, pin)
	if otp == "" {
		return map[string]any{"success": false, "error": "signup pin otp required"}
	}
	if pin == "" {
		return map[string]any{"success": false, "error": "gopay pin missing"}
	}
	client, err := s.newClientWithState(ctx, state, true)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	verificationID := stateString(state, "_signup_pin_verification_id")
	method := stringx.FirstNonEmpty(stateString(state, "_signup_pin_verification_method"), "otp_sms")
	otpToken := stateString(state, "_signup_pin_otp_token")
	if verificationID == "" || otpToken == "" {
		return map[string]any{"success": false, "error": "signup pin otp state missing"}
	}
	verifyResp, err := client.Auth.Post(ctx, "/cvs/v1/verify", s.authBody(map[string]any{
		"data":                map[string]any{"otp": otp, "otp_token": otpToken},
		"flow":                "goto_pin_wa_sms",
		"verification_id":     verificationID,
		"verification_method": method,
	}))
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if verifyResp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("pin otp verify failed", verifyResp)}
	}
	verificationToken := verificationTokenFrom(verifyResp.Data())
	if verificationToken == "" {
		return map[string]any{"success": false, "error": "pin verification_token missing"}
	}
	setupResp, err := client.Customer.Request(ctx, http.MethodPost, "/api/v2/users/pins/setup/tokens", map[string]any{
		"client_id":    stateString(state, "_signup_pin_client_id"),
		"pin":          pin,
		"challenge_id": stateString(state, "_signup_pin_challenge_id"),
	}, http.Header{
		"Verification-Token": []string{"Bearer " + verificationToken},
		"Is-Token-Required":  []string{"false"},
	})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if setupResp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("pin setup failed", setupResp)}
	}
	phone := stringx.FirstNonEmpty(stateString(state, "_signup_phone"), stateString(state, "phone"))
	state["phone"] = phone
	state["stage"] = "ready"
	updatePINSetupState(state, true)
	state["ready_at"] = time.Now().Unix()
	delete(state, "last_error")
	deleteKeys(state, signupAccountStateKeys...)
	deleteKeys(state, signupOTPStateKeys...)
	deleteKeys(state, signupPINStateKeys...)
	return map[string]any{"success": true, "phone": phone, "pin_setup_complete": true}
}
