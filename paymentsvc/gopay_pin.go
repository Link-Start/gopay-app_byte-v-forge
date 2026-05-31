package paymentsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (c *charger) tokenizePIN(ctx context.Context, challengeID, clientID string, linking bool) (string, error) {
	if strings.TrimSpace(c.pin) == "" {
		return "", fmt.Errorf("pin is required")
	}
	headers, err := c.paymentAttemptHeaders(c.gopayPINHeaders(linking))
	if err != nil {
		return "", err
	}
	resp, err := c.paymentHTTP.request(ctx, http.MethodPost, "https://customer.gopayapi.com/api/v1/users/pin/tokens/nb", requestOptions{
		jsonBody: map[string]any{"pin": c.pin, "challenge_id": challengeID, "client_id": clientID},
		headers:  headers,
	})
	if err != nil {
		return "", err
	}
	if resp.status == http.StatusBadRequest || resp.status == http.StatusUnauthorized || resp.status == http.StatusForbidden {
		return "", fmt.Errorf("PIN rejected: %s", resp.excerpt(300))
	}
	if resp.status >= 400 {
		return "", fmt.Errorf("pin tokenize %d: %s", resp.status, resp.excerpt(500))
	}
	token := stringx.FirstNonEmpty(stringAt(resp.data(), "token"), stringAt(resp.data(), "pin_token"), stringAt(resp.json, "data", "token"), stringAt(resp.json, "data", "pin_token"))
	if token == "" {
		return "", fmt.Errorf("pin tokenize: no token in response")
	}
	return token, nil
}
