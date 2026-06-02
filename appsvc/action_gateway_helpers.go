package appsvc

import (
	"strings"

	"github.com/byte-v-forge/common-lib/jsonx"
)

func goPayAppAccountID(value string) string {
	return strings.TrimSpace(value)
}

func normalizeActionOTPChannel(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "wa", "whatsapp", "otp_wa", "gopay_otp_channel_whatsapp":
		return "wa"
	case "sms", "otp_sms", "gopay_otp_channel_sms":
		return "sms"
	default:
		return value
	}
}

func normalizeGoPayWorkflowOperation(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "gopay_account_workflow_operation_")
	switch value {
	case "", "unspecified":
		return "login"
	case "ensure_pin_setup":
		return "ensure_pin_setup"
	case "check_balance":
		return "check_balance"
	case "check_pin":
		return "check_pin"
	case "change_phone":
		return "change_phone"
	case "provision":
		return "provision"
	case "deactivate":
		return "deactivate"
	case "signup":
		return "signup"
	default:
		return value
	}
}

func mapString(data map[string]any, key string) string {
	return strings.TrimSpace(jsonx.StringAt(data, key))
}

func firstNonZeroInt32(values ...int32) int32 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
