package appsvc

import (
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
)

func walletBalance(value any) (int64, string) {
	items, ok := value.([]any)
	if !ok {
		if obj, ok := value.(map[string]any); ok {
			items = []any{obj}
		}
	}
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok || anyString(obj["type"]) != "GOPAY_WALLET" {
			continue
		}
		balance := nestedMap(obj["balance"])
		amount := parseBalanceAmount(balance["value"])
		if amount < 0 {
			amount = parseBalanceAmount(balance["display_value"])
		}
		return amount, stringx.FirstNonEmpty(anyString(balance["currency"]), anyString(obj["currency"]))
	}
	return -1, ""
}

func parseBalanceAmount(value any) int64 {
	if value == nil {
		return -1
	}
	text := anyString(value)
	if text == "" {
		return -1
	}
	var digits strings.Builder
	for _, ch := range text {
		if (ch >= '0' && ch <= '9') || ch == '-' {
			digits.WriteRune(ch)
		}
	}
	raw := digits.String()
	if raw == "" || raw == "-" {
		return -1
	}
	var out int64
	if _, err := fmt.Sscan(raw, &out); err != nil {
		return -1
	}
	return out
}
