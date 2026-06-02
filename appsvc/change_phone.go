package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (s *Server) changePhoneStart(ctx context.Context, state stateMap, pin, newPhone, countryCode string) map[string]any {
	pin = s.resolveGoPayAccountPin(ctx, state, pin)
	cc := phoneCountryCode(s.cfg, countryCode)
	phone := normalizePhoneWithConfig(s.cfg, newPhone, cc)
	if !tokenUsable(state, "token", 30*time.Second) {
		return map[string]any{"success": false, "error": "account token missing"}
	}
	if pin == "" {
		return map[string]any{"success": false, "error": "gopay pin missing"}
	}
	if phone == "" {
		return map[string]any{"success": false, "error": "new_phone required"}
	}
	if stateInt(state, "_temp_phone_usage_"+phone) >= 2 {
		return map[string]any{"success": false, "error": "PHONE_EXHAUSTED"}
	}
	client, err := s.clientForState(ctx, state)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	profile, profileErr := s.loadGojekProfile(ctx, client.Gojek)
	if profileErr == "" {
		s.syncProfileFields(state, profile, cc)
	} else if stateString(state, "email") == "" {
		return map[string]any{"success": false, "error": profileErr}
	}
	body := s.changePhoneProfileBody(state, cc, phone)
	checked := stateString(state, "_checked_change_phone") == phone && stateString(state, "_checked_change_phone_status") == "available"
	if !checked {
		resp, err := client.Gojek.Patch(ctx, "/v5/customers", body)
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}
		if phoneRegisteredResponse(resp) {
			return map[string]any{"success": false, "error": "PHONE_REGISTERED"}
		}
		if otpTokenFrom(resp.Data()) != "" && resp.StatusCode == http.StatusOK {
			return storeChangePhoneOTPState(state, phone, resp.Data())
		}
		if resp.StatusCode != 461 {
			return map[string]any{"success": false, "error": apiError("pin challenge failed", resp)}
		}
	}
	resp, err := client.Gojek.Request(ctx, http.MethodPatch, "/v5/customers", body, http.Header{"pin": []string{pin}})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if phoneRegisteredResponse(resp) {
		return map[string]any{"success": false, "error": "PHONE_REGISTERED"}
	}
	if resp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("pin submit failed", resp)}
	}
	return storeChangePhoneOTPState(state, phone, resp.Data())
}

func storeChangePhoneOTPState(state stateMap, phone string, data any) map[string]any {
	otpToken := otpTokenFrom(data)
	if otpToken == "" {
		return map[string]any{"success": false, "error": "otp_token missing"}
	}
	now := time.Now().Unix()
	state["_change_phone"] = phone
	state["_change_otp_token"] = otpToken
	state["_change_otp_sent_at"] = now
	state["_change_otp_expires_at"] = now + firstNonZero(intForAnyKey(data, "expires_in", "otp_expires_in"), 300)
	state["stage"] = "change_phone_otp_pending"
	deleteKeys(state, "_checked_change_phone", "_checked_change_phone_status", "last_error")
	return map[string]any{"success": true, "new_phone": phone, "otp_sent": true}
}

func (s *Server) changePhoneRetry(ctx context.Context, state stateMap) map[string]any {
	otpToken := stateString(state, "_change_otp_token")
	phone := stateString(state, "_change_phone")
	if otpToken == "" || phone == "" {
		return map[string]any{"success": false, "error": fmt.Sprintf("not waiting for change phone otp: %s", stringx.FirstNonEmpty(stateString(state, "stage"), "idle"))}
	}
	client, err := s.clientForState(ctx, state)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	resp, err := client.Gojek.Post(ctx, "/v2/otp/retry", map[string]any{"otp_token": otpToken, "channel_type": "sms"})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if resp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("otp retry failed", resp)}
	}
	newToken := otpTokenFrom(resp.Data())
	if newToken == "" {
		return map[string]any{"success": false, "error": "retry otp_token missing"}
	}
	now := time.Now().Unix()
	state["_change_otp_token"] = newToken
	state["_change_otp_sent_at"] = now
	state["_change_otp_expires_at"] = now + firstNonZero(intForAnyKey(resp.Data(), "otp_expires_in", "expires_in"), 300)
	state["stage"] = "change_phone_otp_pending"
	delete(state, "last_error")
	return map[string]any{"success": true, "otp_sent": true}
}

func (s *Server) changePhoneComplete(ctx context.Context, state stateMap, otp string) map[string]any {
	otpToken := stateString(state, "_change_otp_token")
	phone := stateString(state, "_change_phone")
	if otpToken == "" || phone == "" {
		return map[string]any{"success": false, "error": fmt.Sprintf("not waiting for change phone otp: %s", stringx.FirstNonEmpty(stateString(state, "stage"), "idle"))}
	}
	if strings.TrimSpace(otp) == "" {
		return map[string]any{"success": false, "error": "otp required"}
	}
	client, err := s.clientForState(ctx, state)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	resp, err := client.Gojek.Post(ctx, "/v5/customers/verificationUpdateProfile", map[string]any{"otp": strings.TrimSpace(otp), "otp_token": otpToken})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if resp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("otp verify failed", resp)}
	}
	if ok, errMsg := s.confirmChangePhone(ctx, client.Gojek, "", phone); !ok {
		state["last_error"] = errMsg
		return map[string]any{"success": false, "error": errMsg}
	}
	state["phone"] = phone
	state["_temp_phone_usage_"+phone] = stateInt(state, "_temp_phone_usage_"+phone) + 1
	s.migrateActiveTokensToTmp(state, phone)
	state["stage"] = "phone_changed"
	deleteKeys(state, "last_error", "_change_phone", "_change_otp_token", "_change_otp_sent_at", "_change_otp_expires_at")
	return map[string]any{"success": true}
}
