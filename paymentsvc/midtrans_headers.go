package paymentsvc

import (
	"encoding/base64"
	"net/http"
)

type midtransHeaderOptions struct {
	jsonBody bool
	source   bool
	auth     bool
	origin   bool
}

func (c *charger) midtransHeaders(snapToken string, opts midtransHeaderOptions) http.Header {
	headers := http.Header{
		"Accept":  []string{"application/json"},
		"Referer": []string{midtransRedirectionURL(snapToken)},
	}
	if opts.jsonBody {
		headers.Set("Content-Type", "application/json")
		opts.origin = true
	}
	if opts.origin {
		headers.Set("Origin", "https://app.midtrans.com")
	}
	if opts.source {
		headers.Set("x-source", "snap")
		headers.Set("x-source-app-type", "redirection")
		headers.Set("x-source-version", "2.3.0")
	}
	if opts.auth {
		headers.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.cfg.MidtransClientID+":")))
	}
	return headers
}

func midtransRedirectionURL(snapToken string) string {
	return "https://app.midtrans.com/snap/v4/redirection/" + snapToken
}
