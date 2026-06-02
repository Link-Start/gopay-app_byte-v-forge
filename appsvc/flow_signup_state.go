package appsvc

import "time"

func (s *Server) persistSignupOTP(state stateMap, verificationID, method, otpToken string) {
	now := time.Now().Unix()
	state["_signup_verification_id"] = verificationID
	state["_signup_verification_method"] = method
	if channel := normalizeActionOTPChannel(method); channel != "" {
		state["_otp_channel"] = channel
	}
	state["_signup_otp_token"] = otpToken
	state["_signup_otp_sent_at"] = now
	state["_signup_otp_expires_at"] = now + int64(s.cfg.OTPTimeout.Seconds())
	state["stage"] = "signup_otp_pending"
	delete(state, "last_error")
}

func (s *Server) clearLoginState(state stateMap, reason string) {
	deleteKeys(state, loginStateKeys...)
	if stage := stateString(state, "stage"); stage == "login" || stage == "login_otp_pending" {
		if stateInt(state, "deactivated_at") > 0 {
			state["stage"] = "deactivated"
		} else {
			state["stage"] = "idle"
		}
	}
	if reason != "" {
		state["last_error"] = reason
	}
}

func (s *Server) clearSignupState(state stateMap, reason string) {
	deleteKeys(state, signupAccountStateKeys...)
	deleteKeys(state, signupOTPStateKeys...)
	deleteKeys(state, signupPINStateKeys...)
	stage := stateString(state, "stage")
	if stage == "signup" || stage == "signup_otp_pending" || stage == "signup_pin_required" || stage == "signup_pin_otp_pending" {
		if stateInt(state, "deactivated_at") > 0 && stateString(state, "token") == "" {
			state["stage"] = "deactivated"
		} else {
			state["stage"] = "idle"
		}
	}
	if reason != "" {
		state["last_error"] = reason
	}
}
