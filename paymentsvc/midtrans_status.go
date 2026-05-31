package paymentsvc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (c *charger) midtransPollStatus(ctx context.Context, snapToken string) (map[string]any, error) {
	var last string
	limit := statusPollLimit
	if c.requiresManualConfirmation() {
		limit = qrisStatusPollLimit
	}
	for range limit {
		resp, err := c.paymentHTTP.request(ctx, http.MethodGet, "https://app.midtrans.com/snap/v1/transactions/"+snapToken+"/status", requestOptions{
			headers: c.midtransHeaders(snapToken, midtransHeaderOptions{source: true}),
		})
		if err != nil {
			last = err.Error()
		} else if resp.status == http.StatusOK {
			status := stringAt(resp.json, "transaction_status")
			statusCode := stringAt(resp.json, "status_code")
			if status == "settlement" || status == "capture" || statusCode == "200" {
				return resp.json, nil
			}
			if status == "deny" || status == "cancel" || status == "expire" || status == "failure" {
				return nil, fmt.Errorf("midtrans transaction failed: %s", resp.excerpt(500))
			}
			last = fmt.Sprintf("status=%q status_code=%q", status, statusCode)
		} else {
			last = fmt.Sprintf("http %d: %s", resp.status, resp.excerpt(150))
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return map[string]any{}, fmt.Errorf("midtrans status poll timeout: %s", last)
}

func (c *charger) followMidtransFinishRedirect(ctx context.Context, state map[string]any, midtransStatus map[string]any) string {
	finishURL := stringx.FirstNonEmpty(stringAt(midtransStatus, "finish_redirect_url"), stringAt(midtransStatus, "finish_200_redirect_url"), stringAt(state, "finish_redirect_url"), stringAt(state, "finish_200_redirect_url"))
	if finishURL == "" {
		return ""
	}
	_, _ = c.paymentHTTP.request(ctx, http.MethodGet, finishURL, requestOptions{
		headers: http.Header{
			"Accept":  []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
			"Referer": []string{midtransRedirectionURL(stringAt(state, "snap_token"))},
		},
	})
	return finishURL
}
