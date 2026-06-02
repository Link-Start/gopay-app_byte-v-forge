package appsvc

import (
	"fmt"
	"time"

	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/common-lib/stringx"
)

func (s *Server) signupCooldownResult(state stateMap) map[string]any {
	until := stateInt(state, "_signup_cooldown_until")
	now := time.Now().Unix()
	if until <= now {
		if until > 0 {
			deleteKeys(
				state,
				"_signup_cooldown_until",
				"_signup_rate_limited_at",
				"_signup_rate_limit_scope",
				"_signup_rate_limit_status",
				"_signup_rate_limit_phone",
				"_signup_rate_limit_country_code",
			)
		}
		return nil
	}
	retryAfter := until - now
	scope := stringx.FirstNonEmpty(stateString(state, "_signup_rate_limit_scope"), "signup")
	state["last_error"] = "SIGNUP_RATE_LIMITED"
	return map[string]any{
		"success":              false,
		"error":                fmt.Sprintf("signup cooldown active: scope=%s retry_after_seconds=%d", scope, retryAfter),
		"rate_limit_scope":     scope,
		"cooldown_until":       until,
		"retry_after_seconds":  retryAfter,
		"cooldown_state_reuse": true,
	}
}

func (s *Server) signupRateLimitResult(state stateMap, scope string, phone string, countryCode string, label string, resp *httpjson.Response) map[string]any {
	now := time.Now()
	cooldown := s.signupRateLimitCooldown(resp)
	until := now.Add(cooldown).Unix()
	state["_signup_rate_limited_at"] = now.Unix()
	state["_signup_rate_limit_scope"] = scope
	state["_signup_rate_limit_status"] = statusCode(resp)
	state["_signup_rate_limit_phone"] = phone
	state["_signup_rate_limit_country_code"] = countryCode
	state["_signup_cooldown_until"] = until
	state["stage"] = "idle"
	state["last_error"] = "SIGNUP_RATE_LIMITED"
	return map[string]any{
		"success":             false,
		"error":               apiError(label, resp),
		"raw_json":            safeJSON(resp.Payload),
		"rate_limit_scope":    scope,
		"cooldown_until":      until,
		"cooldown_seconds":    int64(cooldown.Seconds()),
		"retry_after_seconds": int64(cooldown.Seconds()),
	}
}

func (s *Server) signupRateLimitCooldown(resp *httpjson.Response) time.Duration {
	if retryAfter := retryTimerSeconds(resp); retryAfter > 0 {
		return time.Duration(retryAfter) * time.Second
	}
	if s.cfg.SignupRateLimitCooldown > 0 {
		return s.cfg.SignupRateLimitCooldown
	}
	return 15 * time.Minute
}
