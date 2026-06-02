package appsvc

import (
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (s *Server) persistLoginProbe(state stateMap, phone, countryCode string, data any) (string, []string, string) {
	verificationID := verificationIDFrom(data)
	methods := methodsFrom(data)
	defaultMethod := stringForAnyKey(data, "default_method", "defaultMethod")
	now := time.Now().Unix()
	state["_login_phone"] = phone
	state["_login_country_code"] = countryCode
	state["_login_verification_id"] = verificationID
	state["_login_methods"] = methods
	state["_login_default_method"] = defaultMethod
	state["_login_methods_checked_at"] = now
	state["stage"] = "login"
	delete(state, "last_error")
	return verificationID, methods, defaultMethod
}

func (s *Server) reusableLoginProbe(state stateMap, phone, countryCode string) (string, []string, string, bool) {
	if stateString(state, "_login_phone") != phone || stateString(state, "_login_country_code") != countryCode {
		return "", nil, "", false
	}
	verificationID := stateString(state, "_login_verification_id")
	if verificationID == "" {
		return "", nil, "", false
	}
	checkedAt := stateInt(state, "_login_methods_checked_at")
	if checkedAt <= 0 {
		return "", nil, "", false
	}
	ttl := s.loginProbeTTL()
	if time.Now().Unix() >= checkedAt+int64(ttl.Seconds()) {
		return "", nil, "", false
	}
	return verificationID, methodsFrom(state["_login_methods"]), stateString(state, "_login_default_method"), true
}

func (s *Server) loginProbeTTL() time.Duration {
	ttl := s.cfg.OTPTimeout
	if ttl <= 0 || ttl > 5*time.Minute {
		return 5 * time.Minute
	}
	return ttl
}

func (s *Server) persistLoginReady(state stateMap, tokenData map[string]any, phone string) {
	s.storeTokenResponse(state, tokenData, false)
	state["phone"] = phone
	state["stage"] = "ready"
	state["ready_at"] = time.Now().Unix()
	delete(state, "last_error")
	deleteKeys(state, loginStateKeys...)
}

func (s *Server) persistLoginOTP(state stateMap, phone, countryCode, verificationID, method, otpToken, twoFAToken, flow string) {
	now := time.Now().Unix()
	state["_login_phone"] = phone
	state["_login_country_code"] = countryCode
	state["_login_verification_id"] = verificationID
	state["_login_flow"] = stringx.FirstNonEmpty(flow, "login_2fa")
	state["_login_verification_method"] = method
	if channel := normalizeActionOTPChannel(method); channel != "" {
		state["_otp_channel"] = channel
	}
	state["_login_otp_token"] = otpToken
	state["_login_2fa_token"] = twoFAToken
	state["_login_otp_sent_at"] = now
	state["_login_otp_expires_at"] = now + int64(s.cfg.OTPTimeout.Seconds())
	state["stage"] = "login_otp_pending"
	delete(state, "last_error")
}
