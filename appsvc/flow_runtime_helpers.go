package appsvc

import (
	"errors"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/common-lib/jsonx"
	"github.com/byte-v-forge/common-lib/stringx"

	"github.com/byte-v-forge/gopay-app/otpchannel"
)

func responseShape(resp *httpjson.Response) map[string]any {
	if resp == nil {
		return map[string]any{"status": 0}
	}
	payloadKeys := sortedKeys(resp.Payload)
	data := resp.Data()
	return map[string]any{
		"status":        resp.StatusCode,
		"payload_keys":  payloadKeys,
		"data_keys":     sortedKeys(data),
		"success":       resp.Payload["success"],
		"methods_count": len(methodsFrom(data)),
	}
}

func sortedKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func retryableGoPayTransportError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "eof") ||
		strings.Contains(text, "connection reset") ||
		strings.Contains(text, "connection refused") ||
		strings.Contains(text, "timeout")
}

func loginMethodsRateLimitedError() string {
	return "GoPay login methods still rate limited after identity rotation"
}

func loginMethodsBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return time.Second
	}
	if attempt > 5 {
		attempt = 5
	}
	return time.Duration(attempt) * time.Second
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func otpMethodFromChannel(value string) string {
	return otpchannel.ProviderMethod(value)
}

func chooseOTPMethod(methods []string, preferred, defaultMethod string) string {
	explicit := otpMethodFromChannel(preferred)
	if strings.TrimSpace(preferred) != "" && explicit == "" {
		return ""
	}
	if explicit != "" {
		if len(methods) == 0 || contains(methods, explicit) {
			return explicit
		}
		return ""
	}
	defaultMethod = stringx.FirstNonEmpty(otpMethodFromChannel(defaultMethod), otpchannel.DefaultProviderMethod())
	fallbacks := otpchannel.ProviderMethodFallbacks(defaultMethod)
	for _, method := range fallbacks {
		if method != "" && (len(methods) == 0 || contains(methods, method)) {
			return method
		}
	}
	return defaultMethod
}

func firstAccountID(value any) string {
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return ""
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		return ""
	}
	return stringx.FirstNonEmpty(
		jsonx.StringAt(first, "account_id"),
		jsonx.StringAt(first, "accountId"),
		jsonx.StringAt(first, "id"),
	)
}

func accountListFrom(value any) any {
	if _, ok := value.([]any); ok {
		return value
	}
	if obj, ok := jsonObject(value); ok {
		for _, key := range []string{"account_list", "accountList", "accounts"} {
			if items := obj[key]; items != nil {
				return items
			}
		}
	}
	return nil
}
