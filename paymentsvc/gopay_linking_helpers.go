package paymentsvc

import "strings"

func linkingConsentRequiresOTP(value any) bool {
	found, ok := findBoolField(value, map[string]bool{"otp_required": true, "is_otp_required": true, "requires_otp": true, "need_otp": true, "needs_otp": true})
	if ok {
		return found
	}
	return true
}

func findBoolField(value any, names map[string]bool) (bool, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			if names[strings.ToLower(key)] {
				if b, ok := item.(bool); ok {
					return b, true
				}
			}
			if found, ok := findBoolField(item, names); ok {
				return found, true
			}
		}
	case []any:
		for _, item := range typed {
			if found, ok := findBoolField(item, names); ok {
				return found, true
			}
		}
	}
	return false, false
}
