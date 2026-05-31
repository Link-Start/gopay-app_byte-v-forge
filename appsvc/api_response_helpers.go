package appsvc

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/httpjson"
)

func apiError(label string, resp *httpjson.Response) string {
	if resp == nil {
		return label + ": no response"
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return "AUTH_INVALID"
	}
	return fmt.Sprintf("%s: status %d %s", label, resp.StatusCode, compactErrorDetail(resp.Payload))
}

func responseErrors(resp *httpjson.Response) []any {
	if resp == nil {
		return nil
	}
	for _, source := range []any{resp.Payload["errors"], resp.Data()["errors"]} {
		if items, ok := source.([]any); ok {
			return items
		}
	}
	return nil
}

func responseText(resp *httpjson.Response) string {
	if resp == nil {
		return ""
	}
	return string(resp.Body)
}

func isRateLimited(resp *httpjson.Response) bool {
	if resp == nil {
		return false
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}
	for _, err := range responseErrors(resp) {
		text := strings.ToLower(compactErrorDetail(err))
		if strings.Contains(text, "ratelimited") {
			return true
		}
	}
	return false
}

func loginMethodsInvalidUser(resp *httpjson.Response) bool {
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		return false
	}
	for _, err := range responseErrors(resp) {
		text := strings.ToLower(compactErrorDetail(err))
		if strings.Contains(text, "invalid user") || strings.Contains(text, "could not find the user") {
			return true
		}
	}
	return false
}

func firstNonZero(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
