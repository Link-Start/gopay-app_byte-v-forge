package paymentsvc

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/common-lib/timex"
)

var midtransReferenceRE = regexp.MustCompile(`reference=([a-f0-9-]{36})`)

func (c *charger) startPreparedLinkingUntilOTP(ctx context.Context, state map[string]any, otpChannel string) (map[string]any, error) {
	snapToken := stringAt(state, "snap_token")
	if snapToken == "" {
		return nil, fmt.Errorf("prepared payment is missing snap_token")
	}
	if checkoutURL := stringAt(state, "checkout_url"); checkoutURL != "" {
		c.checkoutURL = checkoutURL
	}
	if c.phone == "" {
		return nil, fmt.Errorf("gopay_phone is required before linking")
	}
	if c.countryCode == "" {
		return nil, fmt.Errorf("gopay_country_code is required before linking")
	}
	return c.startLinkingUntilOTP(ctx, snapToken, stringAt(state, "cs_id"), stringAt(state, "stripe_pk"), otpChannel)
}

func (c *charger) midtransLoadTransaction(ctx context.Context, snapToken string) error {
	_, _ = c.paymentHTTP.request(ctx, http.MethodGet, midtransRedirectionURL(snapToken), requestOptions{
		headers: http.Header{
			"Accept":  []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
			"Referer": []string{"https://pay.openai.com/"},
		},
	})
	resp, err := c.paymentHTTP.request(ctx, http.MethodGet, "https://app.midtrans.com/snap/v1/transactions/"+snapToken, requestOptions{
		headers: c.midtransHeaders(snapToken, midtransHeaderOptions{source: true}),
	})
	if err != nil {
		return err
	}
	if err := resp.require(http.StatusOK, "midtrans transaction"); err != nil {
		return err
	}
	merchantID := stringx.FirstNonEmpty(stringAt(resp.json, "merchant", "merchant_id"), stringAt(resp.json, "merchant", "id"))
	if merchantID != "" {
		c.midtransMerchant = merchantID
	}
	_ = c.midtransWarmSnap(ctx, snapToken)
	return nil
}

func (c *charger) midtransWarmSnap(ctx context.Context, snapToken string) error {
	_, _ = c.paymentHTTP.request(ctx, http.MethodPost, "https://app.midtrans.com/snap/v1/promos/"+snapToken+"/search", requestOptions{
		headers: c.midtransHeaders(snapToken, midtransHeaderOptions{source: true, origin: true}),
	})
	_, _ = c.paymentHTTP.request(ctx, http.MethodGet, "https://app.midtrans.com/snap/v3/experiment", requestOptions{
		query:   mapValues("id", snapToken),
		headers: c.midtransHeaders(snapToken, midtransHeaderOptions{source: true}),
	})
	return nil
}

func (c *charger) midtransInitLinking(ctx context.Context, snapToken string) (string, error) {
	url := "https://app.midtrans.com/snap/v3/accounts/" + snapToken + "/linking"
	body := map[string]any{"type": "gopay", "country_code": c.countryCode, "phone_number": c.phone}
	baseHeaders := c.midtransHeaders(snapToken, midtransHeaderOptions{jsonBody: true})
	authHeaders := c.midtransHeaders(snapToken, midtransHeaderOptions{jsonBody: true, auth: true})
	lastErr := ""
	bypassTried := false
	for range linkRetryLimit + 1 {
		headers, err := c.paymentAttemptHeaders(authHeaders)
		if err != nil {
			return "", err
		}
		resp, err := c.paymentHTTP.request(ctx, http.MethodPost, url, requestOptions{jsonBody: body, headers: headers})
		if err != nil {
			return "", err
		}
		if ref := parseLinkingReference(resp); ref != "" {
			return ref, nil
		}
		if resp.status == http.StatusNotAcceptable {
			lastErr = resp.excerpt(200)
			if err := timex.Sleep(ctx, linkRetrySleep); err != nil {
				return "", err
			}
			continue
		}
		if !bypassTried && linkingRateLimited(resp) {
			bypassTried = true
			bypassHeaders, err := c.paymentAttemptHeaders(baseHeaders)
			if err != nil {
				return "", err
			}
			bypassResp, err := c.paymentHTTP.request(ctx, http.MethodPost, url, requestOptions{jsonBody: body, headers: bypassHeaders})
			if err != nil {
				return "", err
			}
			if ref := parseLinkingReference(bypassResp); ref != "" {
				return ref, nil
			}
			return "", fmt.Errorf("midtrans linking bypass failed status=%d body=%s", bypassResp.status, bypassResp.excerpt(300))
		}
		return "", fmt.Errorf("midtrans linking unexpected status=%d body=%s", resp.status, resp.excerpt(300))
	}
	return "", fmt.Errorf("midtrans linking exhausted retries: %s", lastErr)
}

func parseLinkingReference(resp *httpResult) string {
	if resp == nil || resp.status != http.StatusCreated {
		return ""
	}
	match := midtransReferenceRE.FindStringSubmatch(stringAt(resp.json, "activation_link_url"))
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func linkingRateLimited(resp *httpResult) bool {
	if resp == nil {
		return false
	}
	if resp.status == http.StatusTooManyRequests {
		return true
	}
	text := strings.ToLower(string(resp.body))
	return strings.Contains(text, "technical error") || strings.Contains(text, "too many") || strings.Contains(text, "rate limit") || strings.Contains(text, "rate_limit")
}
