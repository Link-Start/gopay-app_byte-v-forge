package appsvc

import (
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
)

func phoneCountryCode(cfg Config, explicit string) string {
	value := strings.TrimSpace(explicit)
	if value == "" {
		value = "62"
	}
	switch strings.ToUpper(strings.TrimPrefix(value, "+")) {
	case "ID", "IDN", "INDONESIA":
		return "+62"
	}
	if strings.HasPrefix(value, "+") {
		return value
	}
	return "+" + value
}

func normalizePhone(phone string, countryCode string) string {
	prefix := strings.TrimPrefix(phoneCountryCode(Config{}, countryCode), "+")
	value := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	if strings.HasPrefix(value, prefix) {
		return strings.TrimPrefix(value, prefix)
	}
	return value
}

func normalizePhoneWithConfig(cfg Config, phone string, countryCode string) string {
	prefix := strings.TrimPrefix(phoneCountryCode(cfg, countryCode), "+")
	value := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	if strings.HasPrefix(value, prefix) {
		return strings.TrimPrefix(value, prefix)
	}
	return value
}

func digitsOnly(value string) string {
	return stringx.Digits(value)
}
