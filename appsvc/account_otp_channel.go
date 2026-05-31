package appsvc

func gopayAccountOTPChannelFromState(state stateMap) string {
	return normalizeActionOTPChannel(firstNonEmpty(
		stateString(state, "_otp_channel"),
		stateString(state, "otp_channel"),
		stateString(state, "_login_verification_method"),
		stateString(state, "_signup_verification_method"),
		stateString(state, "_signup_pin_verification_method"),
	))
}

func persistGopayAccountOTPChannel(state stateMap) {
	if channel := gopayAccountOTPChannelFromState(state); channel != "" {
		state["_otp_channel"] = channel
	}
}
