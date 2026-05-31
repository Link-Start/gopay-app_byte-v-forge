package paymentsvc

import (
	"context"
	"strings"
)

type charger struct {
	cfg              Config
	paymentProfile   requestProfile
	paymentHTTP      *httpSession
	countryCode      string
	phone            string
	pin              string
	tokenization     string
	amount           int64
	currency         string
	checkoutURL      string
	processorEntity  string
	midtransMerchant string
}

func (s *Server) newCharger(ctx context.Context, input StartInput) (*charger, error) {
	_ = ctx
	paymentProfile := s.cfg.PaymentProfile
	paymentFingerprint := paymentProfile.fingerprint()
	paymentHTTP, err := newHTTPSession(paymentProfile.ProxyURL, paymentFingerprint)
	if err != nil {
		return nil, err
	}
	paymentFingerprint.applyBrowserHeaders(paymentHTTP.headers)
	return &charger{
		cfg:            s.cfg,
		paymentProfile: paymentProfile,
		paymentHTTP:    paymentHTTP,
		countryCode:    normalizeCountryCode(input.CountryCode),
		phone:          normalizeDigits(input.Phone),
		pin:            strings.TrimSpace(input.PIN),
		tokenization:   normalizeTokenization(input.Tokenization),
		amount:         input.Amount,
		currency:       strings.ToUpper(strings.TrimSpace(input.Currency)),
	}, nil
}

func (c *charger) close() {
	if c != nil && c.paymentHTTP != nil {
		c.paymentHTTP.close()
	}
}

func (c *charger) requiresManualConfirmation() bool {
	return requiresManualPaymentConfirmation(c.tokenization)
}

func requiresManualPaymentConfirmation(tokenization string) bool {
	value := strings.ToLower(strings.TrimSpace(tokenization))
	return value == "false" || value == "qris"
}

func isQRISTokenization(tokenization string) bool {
	return strings.EqualFold(strings.TrimSpace(tokenization), "qris")
}

func normalizeTokenization(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultTokenization
	}
	return value
}
