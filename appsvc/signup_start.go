package appsvc

import (
	"context"
	"strings"
)

type signupStartInput struct {
	Phone          string
	CountryCode    string
	Name           string
	Email          string
	OTPChannel     string
	SkipPhoneProbe bool
}

func (s *Server) startSignup(ctx context.Context, state stateMap, phone, name, email, countryCode, otpChannel string, skipPhoneProbe bool) map[string]any {
	input := s.signupStartInput(phone, name, email, countryCode, otpChannel, skipPhoneProbe)
	if input.Phone == "" {
		return map[string]any{"success": false, "error": "signup phone missing"}
	}
	if input.Name == "" {
		return map[string]any{"success": false, "error": "signup name missing"}
	}
	if cooldown := s.signupCooldownResult(state); cooldown != nil {
		return cooldown
	}
	device, proxyURL, err := s.prepareSignupStartState(ctx, state, input)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	client, err := s.newClient(ctx, "", proxyURL, device)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	supportWarmup := s.warmupSignupSupport(ctx, state, client)
	if result := s.probeSignupPhone(ctx, state, client, input, supportWarmup); result != nil {
		return result
	}
	return s.requestSignupOTP(ctx, state, device, proxyURL, input, supportWarmup)
}

func (s *Server) signupStartInput(phone, name, email, countryCode, otpChannel string, skipPhoneProbe bool) signupStartInput {
	cc := phoneCountryCode(s.cfg, countryCode)
	normalized := normalizePhoneWithConfig(s.cfg, phone, cc)
	profileName, profileEmail := s.signupProfile(normalized, name, email)
	return signupStartInput{
		Phone:          normalized,
		CountryCode:    cc,
		Name:           strings.TrimSpace(profileName),
		Email:          strings.TrimSpace(profileEmail),
		OTPChannel:     otpChannel,
		SkipPhoneProbe: skipPhoneProbe,
	}
}
