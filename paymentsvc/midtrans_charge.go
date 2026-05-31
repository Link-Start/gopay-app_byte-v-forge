package paymentsvc

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (c *charger) midtransCreateChargeData(ctx context.Context, snapToken string) (map[string]any, error) {
	resp, err := c.midtransCreateCharge(ctx, snapToken)
	if err != nil {
		return nil, err
	}
	if err := newMidtransChargeDeniedError(resp.json); err != nil {
		return nil, err
	}
	chargeRef := extractMidtransChargeReference(resp.json)
	qrString := stringx.FirstNonEmpty(stringAt(resp.json, "qr_string"), stringAt(resp.json, "qris_string"))
	urls := midtransChargeURLs(resp.json)
	if chargeRef == "" && isQRISTokenization(c.tokenization) && (qrString != "" || urls["qr_code_url"] != "") {
		chargeRef = stringx.FirstNonEmpty(extractReferenceFromText(urls["qr_code_url"]), snapToken)
	}
	if chargeRef == "" {
		return nil, fmt.Errorf("midtrans charge: no reference in response: %s", jsonExcerpt(redactMidtransChargeDebug(resp.json), 500))
	}
	data := map[string]any{"charge_ref": chargeRef, "snap_token": snapToken}
	if qrString != "" {
		data["qr_string"] = qrString
	}
	for key, value := range urls {
		data[key] = value
	}
	return data, nil
}

func (c *charger) midtransCreateCharge(ctx context.Context, snapToken string) (*httpResult, error) {
	url := "https://app.midtrans.com/snap/v2/transactions/" + snapToken + "/charge"
	baseHeaders := c.midtransHeaders(snapToken, midtransHeaderOptions{jsonBody: true, source: true})
	attempts := []map[string]any{{
		"payment_type":  "gopay",
		"tokenization":  c.tokenization,
		"promo_details": nil,
	}}
	if !isQRISTokenization(c.tokenization) {
		base := attempts[0]
		attempts = make([]map[string]any, 0, midtransChargeRetryLimit)
		for range midtransChargeRetryLimit {
			attempts = append(attempts, cloneMap(base))
		}
	}
	if isQRISTokenization(c.tokenization) {
		grossAmount := c.grossAmountString()
		attempts = []map[string]any{{
			"payment_type":  "qris",
			"qris":          map[string]any{"acquirer": "gopay"},
			"gross_amount":  grossAmount,
			"promo_details": nil,
		}, {
			"payment_type":  "gopay",
			"tokenization":  "false",
			"gross_amount":  grossAmount,
			"promo_details": nil,
		}}
	}
	var last *httpResult
	lastErr := ""
	for _, body := range attempts {
		headers, err := c.midtransChargeAttemptHeaders(baseHeaders)
		if err != nil {
			return nil, err
		}
		resp, err := c.paymentHTTP.request(ctx, http.MethodPost, url, requestOptions{jsonBody: body, headers: headers})
		if err != nil {
			return nil, err
		}
		last = resp
		if resp.status == http.StatusOK || resp.status == http.StatusCreated {
			if err := newMidtransChargeDeniedError(resp.json); err != nil {
				return nil, err
			}
			if isQRISTokenization(c.tokenization) && !midtransChargeLooksUsable(resp.json) {
				lastErr = fmt.Sprintf("midtrans charge returned unusable qris response: %s", jsonExcerpt(redactMidtransChargeDebug(resp.json), 500))
				continue
			}
			return resp, nil
		}
	}
	if last == nil {
		return nil, fmt.Errorf("midtrans charge returned empty response")
	}
	if lastErr != "" {
		return nil, fmt.Errorf("%s", lastErr)
	}
	return nil, fmt.Errorf("midtrans charge %d: %s", last.status, last.excerpt(500))
}

func (c *charger) midtransChargeAttemptHeaders(base http.Header) (http.Header, error) {
	headers := cloneHeader(base)
	mergeHeader(headers, c.paymentFingerprint().newAttemptHeaders())
	headers.Set("sec-fetch-dest", "empty")
	headers.Set("sec-fetch-mode", "cors")
	headers.Set("sec-fetch-site", "same-origin")
	return headers, nil
}

func midtransChargeDenial(data map[string]any) string {
	status := stringAt(data, "transaction_status")
	fraud := stringAt(data, "fraud_status")
	if status != "deny" && status != "cancel" && status != "expire" && status != "failure" && fraud != "deny" {
		return ""
	}
	return "midtrans charge denied: " + jsonExcerpt(data, 500)
}

type midtransChargeDeniedError struct {
	data    map[string]any
	message string
}

func (e *midtransChargeDeniedError) Error() string {
	if e == nil {
		return ""
	}
	return e.message
}

func newMidtransChargeDeniedError(data map[string]any) error {
	message := midtransChargeDenial(data)
	if message == "" {
		return nil
	}
	return &midtransChargeDeniedError{data: data, message: message}
}

func (c *charger) grossAmountString() string {
	amount := c.amount
	if amount <= 0 {
		amount = 1
	}
	return strconv.FormatInt(amount, 10)
}
