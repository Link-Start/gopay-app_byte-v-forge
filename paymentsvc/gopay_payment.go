package paymentsvc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/byte-v-forge/common-lib/timex"
)

func (c *charger) gopayPaymentValidate(ctx context.Context, chargeRef string) error {
	query := url.Values{"reference_id": []string{chargeRef}}
	var last *httpResult
	for range 8 {
		headers, err := c.paymentAttemptHeaders(c.gopayHeaders(""))
		if err != nil {
			return err
		}
		resp, err := c.paymentHTTP.request(ctx, http.MethodGet, "https://gwa.gopayapi.com/v1/payment/validate", requestOptions{
			query:   query,
			headers: headers,
		})
		if err != nil {
			return err
		}
		last = resp
		if resp.status == http.StatusOK && boolAt(resp.json, "success") {
			return nil
		}
		if err := timex.Sleep(ctx, 1500*time.Millisecond); err != nil {
			return err
		}
	}
	return fmt.Errorf("payment/validate failed after retries: %d %s", last.status, last.excerpt(250))
}

func (c *charger) gopayPaymentConfirm(ctx context.Context, chargeRef string) (string, string, error) {
	query := url.Values{"reference_id": []string{chargeRef}}
	headers, err := c.paymentAttemptHeaders(c.gopayHeaders(""))
	if err != nil {
		return "", "", err
	}
	resp, err := c.paymentHTTP.request(ctx, http.MethodPost, "https://gwa.gopayapi.com/v1/payment/confirm", requestOptions{
		query:    query,
		jsonBody: map[string]any{"payment_instructions": []any{}},
		headers:  headers,
	})
	if err != nil {
		return "", "", err
	}
	if err := resp.require(http.StatusOK, "payment/confirm"); err != nil {
		return "", "", err
	}
	if !boolAt(resp.json, "success") {
		return "", "", fmt.Errorf("payment/confirm failed: %s", resp.excerpt(500))
	}
	challengeID := stringAt(resp.json, "data", "challenge", "action", "value", "challenge_id")
	clientID := stringAt(resp.json, "data", "challenge", "action", "value", "client_id")
	if challengeID == "" || clientID == "" {
		return "", "", fmt.Errorf("payment/confirm missing challenge")
	}
	return challengeID, clientID, nil
}

func (c *charger) gopayPaymentProcess(ctx context.Context, chargeRef, pinToken string) error {
	query := url.Values{"reference_id": []string{chargeRef}}
	headers, err := c.paymentAttemptHeaders(c.gopayHeaders(""))
	if err != nil {
		return err
	}
	resp, err := c.paymentHTTP.request(ctx, http.MethodPost, "https://gwa.gopayapi.com/v1/payment/process", requestOptions{
		query: query,
		jsonBody: map[string]any{"challenge": map[string]any{
			"type":  "GOPAY_PIN_CHALLENGE",
			"value": map[string]any{"pin_token": pinToken},
		}},
		headers: headers,
	})
	if err != nil {
		return err
	}
	if err := resp.require(http.StatusOK, "payment/process"); err != nil {
		return err
	}
	if !boolAt(resp.json, "success") || stringAt(resp.json, "data", "next_action") != "payment-success" {
		return fmt.Errorf("payment/process failed: %s", resp.excerpt(500))
	}
	return nil
}
