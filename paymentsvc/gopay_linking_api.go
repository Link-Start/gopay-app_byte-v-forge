package paymentsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func (c *charger) gopayValidateReference(ctx context.Context, referenceID string) error {
	headers, err := c.paymentAttemptHeaders(c.gopayHeaders(""))
	if err != nil {
		return err
	}
	resp, err := c.paymentHTTP.request(ctx, http.MethodPost, "https://gwa.gopayapi.com/v1/linking/validate-reference", requestOptions{
		jsonBody: map[string]any{"reference_id": referenceID},
		headers:  headers,
	})
	if err != nil {
		return err
	}
	if err := resp.require(http.StatusOK, "validate-reference"); err != nil {
		return err
	}
	if !boolAt(resp.json, "success") {
		return fmt.Errorf("validate-reference failed: %s", resp.excerpt(500))
	}
	return nil
}

func (c *charger) gopayLinkingConsent(ctx context.Context, referenceID string) (map[string]any, error) {
	headers, err := c.paymentAttemptHeaders(c.gopayHeaders(c.paymentProfile.Locale))
	if err != nil {
		return nil, err
	}
	resp, err := c.paymentHTTP.request(ctx, http.MethodPost, "https://gwa.gopayapi.com/v1/linking/user-consent", requestOptions{
		jsonBody: map[string]any{"reference_id": referenceID},
		headers:  headers,
	})
	if err != nil {
		return nil, err
	}
	if err := resp.require(http.StatusOK, "user-consent"); err != nil {
		return nil, err
	}
	if !boolAt(resp.json, "success") {
		return nil, fmt.Errorf("user-consent failed: %s", resp.excerpt(500))
	}
	return resp.json, nil
}

func (c *charger) gopayResendOTP(ctx context.Context, referenceID string) (map[string]any, error) {
	headers, err := c.paymentAttemptHeaders(c.gopayHeaders(c.paymentProfile.Locale))
	if err != nil {
		return nil, err
	}
	resp, err := c.paymentHTTP.request(ctx, http.MethodPost, "https://gwa.gopayapi.com/v1/linking/resend-otp", requestOptions{
		jsonBody: map[string]any{"reference_id": referenceID},
		headers:  headers,
	})
	if err != nil {
		return nil, err
	}
	if resp.status < 200 || resp.status >= 300 {
		return nil, fmt.Errorf("resend-otp %d: %s", resp.status, resp.excerpt(300))
	}
	return resp.json, nil
}

func (c *charger) gopayValidateOTP(ctx context.Context, referenceID, otp string) (string, string, error) {
	headers, err := c.paymentAttemptHeaders(c.gopayHeaders(c.paymentProfile.Locale))
	if err != nil {
		return "", "", err
	}
	resp, err := c.paymentHTTP.request(ctx, http.MethodPost, "https://gwa.gopayapi.com/v1/linking/validate-otp", requestOptions{
		jsonBody: map[string]any{"reference_id": referenceID, "otp": strings.TrimSpace(otp)},
		headers:  headers,
	})
	if err != nil {
		return "", "", err
	}
	if resp.status != http.StatusOK {
		return "", "", fmt.Errorf("validate-otp %d: %s", resp.status, resp.excerpt(400))
	}
	if !boolAt(resp.json, "success") {
		return "", "", fmt.Errorf("validate-otp failed: %s", resp.excerpt(400))
	}
	challengeID := stringAt(resp.json, "data", "challenge", "action", "value", "challenge_id")
	clientID := stringAt(resp.json, "data", "challenge", "action", "value", "client_id")
	if challengeID == "" || clientID == "" {
		return "", "", fmt.Errorf("validate-otp: missing challenge details")
	}
	return challengeID, clientID, nil
}

func (c *charger) gopayValidatePIN(ctx context.Context, referenceID, pinToken string) error {
	headers, err := c.paymentAttemptHeaders(c.gopayHeaders(c.paymentProfile.Locale))
	if err != nil {
		return err
	}
	resp, err := c.paymentHTTP.request(ctx, http.MethodPost, "https://gwa.gopayapi.com/v1/linking/validate-pin", requestOptions{
		jsonBody: map[string]any{"reference_id": referenceID, "token": pinToken},
		headers:  headers,
	})
	if err != nil {
		return err
	}
	if err := resp.require(http.StatusOK, "validate-pin"); err != nil {
		return err
	}
	if !boolAt(resp.json, "success") {
		return fmt.Errorf("validate-pin failed: %s", resp.excerpt(500))
	}
	return nil
}
