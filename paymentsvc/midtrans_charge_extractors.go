package paymentsvc

import (
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
)

func extractMidtransURL(data map[string]any, names ...string) string {
	wanted := map[string]bool{}
	for _, name := range names {
		wanted[strings.ToLower(name)] = true
		if value := stringAt(data, name); value != "" {
			return value
		}
	}
	items, _ := data["actions"].([]any)
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := strings.ToLower(stringAt(obj, "name"))
		if wanted[name] || (strings.Contains(name, "qr") && (wanted["qr_code_url"] || wanted["qr_code"] || wanted["qrcode"])) {
			if value := stringAt(obj, "url"); value != "" {
				return value
			}
		}
	}
	return ""
}

func midtransChargeURLs(data map[string]any) map[string]string {
	return map[string]string{
		"deeplink_url":            stringx.FirstNonEmpty(extractMidtransURL(data, "deeplink_url", "deeplink"), stringAt(data, "gopay_deeplink_url")),
		"qr_code_url":             stringx.FirstNonEmpty(extractMidtransURL(data, "qr_code_url", "qr_code", "qrcode"), stringAt(data, "qr_string"), stringAt(data, "qris_string"), stringAt(data, "qris_url"), stringAt(data, "gopay_verification_link_url")),
		"finish_redirect_url":     extractMidtransURL(data, "finish_redirect_url"),
		"finish_200_redirect_url": extractMidtransURL(data, "finish_200_redirect_url"),
	}
}

func midtransChargeLooksUsable(data map[string]any) bool {
	if extractMidtransChargeReference(data) != "" {
		return true
	}
	if stringx.FirstNonEmpty(stringAt(data, "qr_string"), stringAt(data, "qris_string")) != "" {
		return true
	}
	urls := midtransChargeURLs(data)
	return urls["qr_code_url"] != "" || urls["deeplink_url"] != ""
}

func redactMidtransChargeDebug(data map[string]any) map[string]any {
	out := map[string]any{}
	for _, key := range []string{"transaction_id", "order_id", "payment_type", "transaction_status", "status_code", "fraud_status", "expiry_time"} {
		if value := stringAt(data, key); value != "" {
			out[key] = value
		}
	}
	for _, key := range []string{"qr_string", "qris_string"} {
		if stringAt(data, key) != "" {
			out[key] = "<present>"
		}
	}
	if actions, ok := data["actions"].([]any); ok {
		items := make([]map[string]string, 0, len(actions))
		for _, item := range actions {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}
			items = append(items, map[string]string{
				"name":   stringAt(obj, "name"),
				"method": stringAt(obj, "method"),
				"url":    stringAt(obj, "url"),
			})
		}
		if len(items) > 0 {
			out["actions"] = items
		}
	}
	return out
}
