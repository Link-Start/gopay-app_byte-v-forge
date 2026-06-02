package appsvc

import (
	"context"
	"net/http"
	"time"

	gopayapp "github.com/byte-v-forge/gopay-app/protocol/app"
)

func (s *Server) warmupSignupSupport(ctx context.Context, state stateMap, client *gopayapp.ClientSet) map[string]any {
	out := map[string]any{"attempted": true}
	if warmupResp, err := client.Customer.InitiateSupportCustomer(ctx); err != nil {
		out["success"] = false
		out["error"] = err.Error()
		state["_signup_support_warmup_error"] = err.Error()
	} else {
		status := 0
		if warmupResp != nil {
			status = warmupResp.StatusCode
		}
		out["success"] = status >= 200 && status < 300
		out["status_code"] = status
		state["_signup_support_warmup_status"] = status
		delete(state, "_signup_support_warmup_error")
	}
	state["_signup_support_warmup_at"] = time.Now().Unix()
	return out
}

func (s *Server) probeSignupPhone(ctx context.Context, state stateMap, client *gopayapp.ClientSet, input signupStartInput, supportWarmup map[string]any) map[string]any {
	if input.SkipPhoneProbe {
		state["_signup_phone_probe_skipped"] = true
		supportWarmup["phone_probe_skipped"] = true
		return nil
	}
	probeResp, err := client.Auth.Post(ctx, "/goto-auth/login/methods", signupProbeBody{
		PhoneNumber:               input.Phone,
		CountryCode:               input.CountryCode,
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
		return s.signupRateLimitResult(state, signupRateLimitScopeProbe, input.Phone, input.CountryCode, rateLimitLabel(signupRateLimitScopeProbe), probeResp)
	}
	if !loginMethodsInvalidUser(probeResp) && probeResp.StatusCode >= http.StatusBadRequest {
		return map[string]any{"success": false, "error": apiError("signup phone probe failed", probeResp), "support_warmup": supportWarmup, "raw_json": safeJSON(probeResp.Payload)}
	}
	return nil
}
