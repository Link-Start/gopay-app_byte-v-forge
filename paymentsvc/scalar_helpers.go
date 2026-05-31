package paymentsvc

import "strings"

func configBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func mapBoolString(ok bool, yes, no string) string {
	if ok {
		return yes
	}
	return no
}

func truncateString(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit]
}
