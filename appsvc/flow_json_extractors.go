package appsvc

import (
	"strings"

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
	if obj, ok := jsonObject(value); ok {
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

func intForAnyKey(value any, keys ...string) int64 {
	wanted := map[string]struct{}{}
	for _, key := range keys {
		wanted[jsonx.NormalizeKey(key)] = struct{}{}
	}
	var walk func(any) int64
	walk = func(current any) int64 {
		if obj, ok := jsonObject(current); ok {
			for key, item := range obj {
				if _, ok := wanted[jsonx.NormalizeKey(key)]; ok {
					if parsed := anyInt(item); parsed != 0 {
						return parsed
					}
				}
			}
			for _, item := range obj {
				if parsed := walk(item); parsed != 0 {
					return parsed
				}
			}
		}
		switch typed := current.(type) {
		case []any:
			for _, item := range typed {
				if parsed := walk(item); parsed != 0 {
					return parsed
				}
			}
		}
		return 0
	}
	return walk(value)
}

func boolForAnyKey(value any, keys ...string) bool {
	wanted := map[string]struct{}{}
	for _, key := range keys {
		wanted[jsonx.NormalizeKey(key)] = struct{}{}
	}
	var walk func(any) bool
	walk = func(current any) bool {
		if obj, ok := jsonObject(current); ok {
			for key, item := range obj {
				if _, ok := wanted[jsonx.NormalizeKey(key)]; ok {
					return anyBool(item)
				}
			}
			for _, item := range obj {
				if walk(item) {
					return true
				}
			}
		}
		if items, ok := current.([]any); ok {
			for _, item := range items {
				if walk(item) {
					return true
				}
			}
		}
		return false
	}
	return walk(value)
}

func stringForAnyKey(value any, keys ...string) string {
	wanted := map[string]struct{}{}
	for _, key := range keys {
		wanted[jsonx.NormalizeKey(key)] = struct{}{}
	}
	var walk func(any) string
	walk = func(current any) string {
		if obj, ok := jsonObject(current); ok {
			for key, item := range obj {
				if _, ok := wanted[jsonx.NormalizeKey(key)]; ok {
					if text := anyString(item); text != "" {
						return text
					}
				}
			}
			for _, item := range obj {
				if text := walk(item); text != "" {
					return text
				}
			}
			return ""
		}
		switch typed := current.(type) {
		case []any:
			for _, item := range typed {
				if text := walk(item); text != "" {
					return text
				}
			}
		}
		return ""
	}
	return walk(value)
}

func jsonObject(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, typed != nil
	case jsonx.Map:
		return map[string]any(typed), typed != nil
	default:
		return nil, false
	}
}

func verificationScopedID(value any) string {
	var walk func(any, bool) string
	walk = func(current any, inVerificationScope bool) string {
		if obj, ok := jsonObject(current); ok {
			for key, item := range obj {
				normalized := jsonx.NormalizeKey(key)
				nextScope := inVerificationScope || strings.Contains(normalized, "verification")
				if nextScope && normalized == "id" {
					if text := anyString(item); text != "" {
						return text
					}
				}
				if text := walk(item, nextScope); text != "" {
					return text
				}
			}
			return ""
		}
		switch typed := current.(type) {
		case []any:
			for _, item := range typed {
				if text := walk(item, inVerificationScope); text != "" {
					return text
				}
			}
		}
		return ""
	}
	return walk(value, false)
}
