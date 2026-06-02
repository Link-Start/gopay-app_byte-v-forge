package appsvc

import (
	"github.com/byte-v-forge/common-lib/jsonx"
	"github.com/byte-v-forge/common-lib/stringx"
)

func stringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := anyString(item); text != "" {
				out = append(out, text)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := anyString(item); text != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func methodsFrom(value any) []string {
	if methods := stringSlice(value); len(methods) > 0 {
		return methods
	}
	if obj, ok := jsonx.Object(value); ok {
		for key, item := range obj {
			if jsonx.NormalizeKey(key) == "methods" {
				if methods := stringSlice(item); len(methods) > 0 {
					return methods
				}
			}
		}
		for _, item := range obj {
			if methods := methodsFrom(item); len(methods) > 0 {
				return methods
			}
		}
		return nil
	}
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			if methods := methodsFrom(item); len(methods) > 0 {
				return methods
			}
		}
	}
	return nil
}

func verificationIDFrom(value any) string {
	if text := stringForAnyKey(value, "verification_id", "verificationId"); text != "" {
		return text
	}
	return verificationScopedID(value)
}

func challengeIDFrom(value any) string {
	return stringx.FirstNonEmpty(
		jsonx.StringAt(value, "challenge_id"),
		jsonx.StringAt(value, "challenge", "action", "value", "challenge_id"),
		jsonx.StringAt(value, "challenge", "value", "challenge_id"),
		stringForAnyKey(value, "challenge_id", "challengeId"),
	)
}

func clientIDFrom(value any) string {
	return stringx.FirstNonEmpty(
		jsonx.StringAt(value, "client_id"),
		jsonx.StringAt(value, "challenge", "action", "value", "client_id"),
		jsonx.StringAt(value, "challenge", "value", "client_id"),
		stringForAnyKey(value, "client_id", "clientId"),
	)
}

func otpTokenFrom(value any) string {
	return stringForAnyKey(value, "otp_token", "otpToken")
}

func verificationTokenFrom(value any) string {
	return stringForAnyKey(value, "verification_token", "verificationToken")
}

func oneFATokenFrom(value any) string {
	return stringForAnyKey(value, "1fa_token", "one_fa_token", "oneFaToken")
}

func twoFATokenFrom(value any) string {
	return stringForAnyKey(value, "2fa_token", "two_fa_token", "twoFaToken")
}
