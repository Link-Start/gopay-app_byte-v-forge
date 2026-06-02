package appsvc

import "time"

func (s *Server) expireLoginIfNeeded(state stateMap) bool {
	if stateString(state, "stage") != "login_otp_pending" {
		return false
	}
	now := time.Now().Unix()
	expiresAt := stateInt(state, "_login_otp_expires_at")
	if expiresAt > 0 && now < expiresAt {
		return false
	}
	if expiresAt == 0 {
		sentAt := stateInt(state, "_login_otp_sent_at")
		if sentAt > 0 && now < sentAt+int64(s.cfg.OTPTimeout.Seconds()) {
			return false
		}
	}
	s.clearLoginState(state, "LOGIN_OTP_TIMEOUT")
	return true
}

func (s *Server) expireSignupIfNeeded(state stateMap) bool {
	stage := stateString(state, "stage")
	now := time.Now().Unix()
	if stage == "signup_otp_pending" && pendingExpired(now, stateInt(state, "_signup_otp_sent_at"), stateInt(state, "_signup_otp_expires_at"), s.cfg.OTPTimeout) {
		deleteKeys(state, signupOTPStateKeys...)
		state["stage"] = "idle"
		state["last_error"] = "SIGNUP_OTP_TIMEOUT"
		return true
	}
	if stage == "signup_pin_otp_pending" && pendingExpired(now, stateInt(state, "_signup_pin_otp_sent_at"), stateInt(state, "_signup_pin_otp_expires_at"), s.cfg.OTPTimeout) {
		deleteKeys(state, signupPINStateKeys...)
		if stateString(state, "token") != "" {
			state["stage"] = "signup_pin_required"
		} else {
			state["stage"] = "idle"
		}
		state["last_error"] = "SIGNUP_PIN_OTP_TIMEOUT"
		return true
	}
	return false
}

func pendingExpired(now, sentAt, expiresAt int64, timeout time.Duration) bool {
	if expiresAt > 0 {
		return now >= expiresAt
	}
	if sentAt > 0 {
		return now >= sentAt+int64(timeout.Seconds())
	}
	return true
}
