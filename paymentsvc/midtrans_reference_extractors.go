package paymentsvc

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	referenceQueryRE           = regexp.MustCompile(`(?:[?&#]|^)(?:reference|reference_id|referenceId)=([A-Za-z0-9-]+)`)
	qrisPathReferenceRE        = regexp.MustCompile(`/qris/[A-Za-z0-9_-]+/([A-Za-z0-9-]+)/qr-code(?:[/?#]|$)`)
	gopayPathReferenceRE       = regexp.MustCompile(`/gopay/([A-Za-z0-9-]+)/qr-code(?:[/?#]|$)`)
	midtransPaymentRefRE       = regexp.MustCompile(`A[0-9]{12,}[A-Za-z0-9]+ID`)
	genericReferenceValueRE    = regexp.MustCompile(`^[A-Za-z0-9-]{6,}$`)
	midtransReferenceQueryKeys = []string{"reference", "reference_id", "referenceId", "tref"}
)

func extractReferenceFromText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if parsed, err := url.Parse(text); err == nil {
		query := parsed.Query()
		for _, key := range midtransReferenceQueryKeys {
			for _, item := range query[key] {
				if value := strings.TrimSpace(item); value != "" {
					return value
				}
			}
		}
	}
	match := referenceQueryRE.FindStringSubmatch(text)
	if len(match) > 1 {
		return match[1]
	}
	for _, pattern := range []*regexp.Regexp{qrisPathReferenceRE, gopayPathReferenceRE} {
		match := pattern.FindStringSubmatch(text)
		if len(match) > 1 {
			return match[1]
		}
	}
	return ""
}

func extractMidtransPaymentRefFromText(text string) string {
	return midtransPaymentRefRE.FindString(strings.TrimSpace(text))
}

func findMidtransPaymentRef(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		for _, item := range typed {
			if found := findMidtransPaymentRef(item); found != "" {
				return found
			}
		}
	case []any:
		for _, item := range typed {
			if found := findMidtransPaymentRef(item); found != "" {
				return found
			}
		}
	case string:
		return extractMidtransPaymentRefFromText(typed)
	}
	return ""
}

func extractMidtransChargeReference(data any) string {
	if ref := findMidtransPaymentRef(data); ref != "" {
		return ref
	}
	if obj, ok := data.(map[string]any); ok {
		for _, key := range []string{"transaction_id", "charge_ref", "reference_id", "reference", "payment_id", "order_id"} {
			if value := strings.TrimSpace(stringAt(obj, key)); value != "" {
				return value
			}
		}
	}
	var walk func(any, string) string
	walk = func(value any, path string) string {
		switch typed := value.(type) {
		case map[string]any:
			for key, item := range typed {
				if found := walk(item, path+"."+key); found != "" {
					return found
				}
			}
		case []any:
			for _, item := range typed {
				if found := walk(item, path); found != "" {
					return found
				}
			}
		case string:
			if reference := extractReferenceFromText(typed); reference != "" {
				return reference
			}
			if strings.Contains(strings.ToLower(path), "reference") && genericReferenceValueRE.MatchString(strings.TrimSpace(typed)) {
				return strings.TrimSpace(typed)
			}
		}
		return ""
	}
	return walk(data, "")
}
