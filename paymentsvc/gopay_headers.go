package paymentsvc

import (
	"net/http"

	"github.com/google/uuid"
)

func (c *charger) gopayHeaders(locale string) http.Header {
	headers := http.Header{
		"Accept":  []string{"application/json, text/plain, */*"},
		"Origin":  []string{"https://merchants-gws-app.gopayapi.com"},
		"Referer": []string{"https://merchants-gws-app.gopayapi.com/"},
	}
	headers.Set("Content-Type", "application/json")
	if locale != "" {
		headers.Set("x-user-locale", locale)
	}
	return headers
}

func (c *charger) gopayPINHeaders(linking bool) http.Header {
	headers := c.gopayHeaders("")
	if linking {
		headers.Set("Origin", "https://pin-web-client.gopayapi.com")
		headers.Set("Referer", "https://pin-web-client.gopayapi.com/")
		headers.Set("x-user-locale", c.paymentProfile.PINLocale)
		headers.Set("x-appversion", "1.0.0")
		headers.Set("x-is-mobile", "false")
		headers.Set("x-platform", c.paymentProfile.Platform)
	}
	headers.Set("x-request-id", uuid.NewString())
	return headers
}

func (c *charger) paymentAttemptHeaders(base http.Header) (http.Header, error) {
	headers := cloneHeader(base)
	mergeHeader(headers, c.paymentFingerprint().newAttemptHeaders())
	headers.Set("sec-fetch-dest", "empty")
	headers.Set("sec-fetch-mode", "cors")
	headers.Set("sec-fetch-site", "same-origin")
	return headers, nil
}

func (c *charger) paymentFingerprint() browserFingerprint {
	if c != nil && c.paymentHTTP != nil {
		return c.paymentHTTP.fingerprint.withFallback(c.paymentProfile.Locale)
	}
	return defaultRequestProfile("payment").fingerprint()
}
