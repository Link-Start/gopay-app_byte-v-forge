package appsvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (s *Server) resolvePaymentAccount(ctx context.Context, req *pb.StartGopayPaymentRequest) (resolvedPaymentAccount, error) {
	key, err := NormalizeGopayAccountID(req.GetGopayAccountId())
	if err != nil {
		return resolvedPaymentAccount{}, err
	}
	state, err := s.loadGopayAccountState(ctx, key)
	if err != nil {
		return resolvedPaymentAccount{}, err
	}
	profile, err := s.loadGopayAccountProfile(ctx, key)
	if err != nil {
		return resolvedPaymentAccount{}, err
	}
	phone := stringx.FirstNonEmpty(stateString(state, "phone"), stateString(state, "_login_phone"), stateString(state, "_signup_phone"), stateString(profile, "phone"), stateString(profile, "wa_phone"))
	country := stringx.FirstNonEmpty(stateString(profile, "country_code"), stateCountryCode(state), "62")
	otpChannel := normalizeActionOTPChannel(stringx.FirstNonEmpty(
		stateString(profile, "otp_channel"),
		gopayAccountOTPChannelFromState(state),
		"sms",
	))
	if paymentRequiresAccountLinking(req.GetTokenization()) {
		if strings.TrimSpace(phone) == "" {
			return resolvedPaymentAccount{}, fmt.Errorf("gopay account phone is required")
		}
		if strings.TrimSpace(country) == "" {
			return resolvedPaymentAccount{}, fmt.Errorf("gopay account country_code is required")
		}
		if strings.TrimSpace(stateString(profile, "pin")) == "" {
			return resolvedPaymentAccount{}, fmt.Errorf("gopay account PIN is not configured")
		}
	}
	state["_otp_channel"] = otpChannel
	if _, err := s.store.SaveAccount(ctx, key, stateJSON(state)); err != nil {
		return resolvedPaymentAccount{}, err
	}
	if stateString(profile, "otp_channel") == "" {
		profile["otp_channel"] = otpChannel
		profile["updated_at_unix"] = int64Now()
		_, _ = s.store.Save(ctx, gopayAccountProfileKey(key), stateJSON(profile))
	}
	return resolvedPaymentAccount{accountID: key, phone: phone, country: country, otpChannel: otpChannel, otpTarget: stringx.FirstNonEmpty(phone, key), pin: stateString(profile, "pin"), state: state}, nil
}

func (s *Server) ensurePaymentBalance(ctx context.Context, account resolvedPaymentAccount, amount int64, currency string) error {
	check := s.checkTokenValid(ctx, account.state)
	if _, err := s.store.SaveAccount(ctx, account.accountID, stateJSON(account.state)); err != nil {
		return err
	}
	if !s.tokenCheckValid(check) {
		return fmt.Errorf("gopay account token is not ready: %s", s.tokenCheckError(check))
	}
	if amount <= 0 {
		return nil
	}
	balance := anyInt(check["balance_amount"])
	balanceCurrency := strings.ToUpper(stringx.FirstNonEmpty(anyString(check["balance_currency"]), "IDR"))
	if currency != "" && balanceCurrency != "" && !strings.EqualFold(balanceCurrency, currency) {
		return fmt.Errorf("gopay balance currency mismatch: %s != %s", balanceCurrency, currency)
	}
	if balance < amount {
		return fmt.Errorf("insufficient gopay balance: %d %s < %d %s", balance, balanceCurrency, amount, stringx.FirstNonEmpty(currency, balanceCurrency))
	}
	if stateString(account.state, "last_error") == "INSUFFICIENT_GOPAY_BALANCE" {
		delete(account.state, "last_error")
		_, _ = s.store.SaveAccount(ctx, account.accountID, stateJSON(account.state))
	}
	return nil
}

func paymentRequiresAccountLinking(tokenization string) bool {
	value := strings.ToLower(strings.TrimSpace(tokenization))
	return value == "" || value == "true"
}

func int64Now() int64 {
	return time.Now().Unix()
}
