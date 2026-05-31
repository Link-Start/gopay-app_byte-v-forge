package paymentsvc

import "strings"

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return strings.TrimSpace(err.Error())
}

func truncateError(err error) string {
	text := errorText(err)
	if len(text) > 500 {
		return text[:500]
	}
	return text
}
