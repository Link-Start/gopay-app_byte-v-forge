package paymentsvc

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/byte-v-forge/common-lib/jsonx"
	"github.com/byte-v-forge/common-lib/jwtx"
	"github.com/byte-v-forge/common-lib/redactx"
	"github.com/byte-v-forge/common-lib/stringx"
)

func stringAt(value any, path ...string) string {
	return strings.TrimSpace(jsonx.StringAt(value, path...))
}

func boolAt(value any, path ...string) bool {
	current := value
	for _, key := range path {
		obj, ok := current.(map[string]any)
		if !ok {
			return false
		}
		current = obj[key]
	}
	switch typed := current.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return false
	}
}

func intAt(value any, path ...string) int64 {
	text := stringAt(value, path...)
	if text == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0
	}
	return int64(parsed)
}

func normalizeDigits(value string) string {
	return stringx.Digits(value)
}

func normalizeCountryCode(value string) string {
	return normalizeDigits(value)
}

func decodeJWTPayload(token string) map[string]any {
	return jwtx.PayloadOrNil(token)
}

func jsonExcerpt(value any, limit int) string {
	raw, err := jsonx.Compact(value)
	if err != nil {
		return redactx.Snippet(redactx.Text(fmt.Sprint(value)), limit)
	}
	return redactx.Snippet(redactx.Text(string(raw)), limit)
}

func mapValues(key, value string) url.Values {
	values := url.Values{}
	values.Set(key, value)
	return values
}
