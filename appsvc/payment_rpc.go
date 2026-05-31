package appsvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/gopay-app/paymentsvc"
	"github.com/byte-v-forge/gopay-app/pb"
)

type resolvedPaymentAccount struct {
	accountID  string
	phone      string
	country    string
	otpChannel string
	otpTarget  string
	pin        string
	state      stateMap
}

func (s *Server) StartGopayPayment(ctx context.Context, req *pb.StartGopayPaymentRequest) (*pb.StartGopayPaymentResponse, error) {
	if s == nil || s.payment == nil {
		return &pb.StartGopayPaymentResponse{Success: false, ErrorMessage: "gopay payment runtime unavailable"}, nil
	}
	if req == nil {
		req = &pb.StartGopayPaymentRequest{}
	}
	account, err := s.resolvePaymentAccount(ctx, req)
	if err != nil {
		return &pb.StartGopayPaymentResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	amount := req.GetAmount()
	currency := strings.ToUpper(firstNonEmpty(req.GetCurrency(), stateString(account.state, "balance_currency"), "IDR"))
	if err := s.ensurePaymentBalance(ctx, account, amount, currency); err != nil {
		return &pb.StartGopayPaymentResponse{Success: false, ErrorMessage: err.Error(), GopayAccountId: account.accountID, Amount: amount, Currency: currency, OtpChannel: account.otpChannel, OtpTarget: account.otpTarget}, nil
	}
	input := paymentsvc.StartInput{
		SnapToken:         req.GetSnapToken(),
		CheckoutURL:       req.GetCheckoutUrl(),
		CheckoutSessionID: req.GetCheckoutSessionId(),
		GopayAccountID:    account.accountID,
		Phone:             account.phone,
		CountryCode:       account.country,
		PIN:               account.pin,
		OTPChannel:        account.otpChannel,
		OTPTarget:         account.otpTarget,
		Tokenization:      req.GetTokenization(),
		Amount:            amount,
		Currency:          currency,
	}
	resp, err := s.payment.StartPayment(ctx, input)
	if resp != nil {
		resp.GopayAccountId = account.accountID
		resp.Amount = amount
		resp.Currency = currency
		resp.OtpChannel = account.otpChannel
		resp.OtpTarget = account.otpTarget
	}
	return resp, err
}

func (s *Server) CompleteGopayPayment(ctx context.Context, req *pb.CompleteGopayPaymentRequest) (*pb.GopayPaymentResponse, error) {
	if s == nil || s.payment == nil {
		return &pb.GopayPaymentResponse{Success: false, ErrorMessage: "gopay payment runtime unavailable"}, nil
	}
	return s.payment.CompleteGopayPayment(ctx, req)
}

func (s *Server) ResendGopayPaymentOTP(ctx context.Context, req *pb.ResendGopayPaymentOTPRequest) (*pb.ResendGopayPaymentOTPResponse, error) {
	if s == nil || s.payment == nil {
		return &pb.ResendGopayPaymentOTPResponse{Success: false, ErrorMessage: "gopay payment runtime unavailable"}, nil
	}
	return s.payment.ResendGopayPaymentOTP(ctx, req)
}

func (s *Server) ConfirmGopayPayment(ctx context.Context, req *pb.ConfirmGopayPaymentRequest) (*pb.GopayPaymentResponse, error) {
	if s == nil || s.payment == nil {
		return &pb.GopayPaymentResponse{Success: false, ErrorMessage: "gopay payment runtime unavailable"}, nil
	}
	return s.payment.ConfirmGopayPayment(ctx, req)
}

func (s *Server) CancelGopayPayment(ctx context.Context, req *pb.CancelGopayPaymentRequest) (*pb.CancelGopayPaymentResponse, error) {
	if s == nil || s.payment == nil {
		return &pb.CancelGopayPaymentResponse{Success: false, ErrorMessage: "gopay payment runtime unavailable"}, nil
	}
	return s.payment.CancelGopayPayment(ctx, req)
}

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
	phone := firstNonEmpty(stateString(state, "phone"), stateString(state, "_login_phone"), stateString(state, "_signup_phone"), stateString(profile, "phone"), stateString(profile, "wa_phone"))
	country := firstNonEmpty(stateString(profile, "country_code"), stateCountryCode(state), "62")
	otpChannel := normalizeActionOTPChannel(firstNonEmpty(
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
	state["_payment_tokenization"] = firstNonEmpty(req.GetTokenization(), "true")
	state["_payment_amount"] = req.GetAmount()
	state["_payment_currency"] = strings.ToUpper(firstNonEmpty(req.GetCurrency(), stateString(state, "balance_currency"), "IDR"))
	if _, err := s.store.SaveAccount(ctx, key, stateJSON(state)); err != nil {
		return resolvedPaymentAccount{}, err
	}
	if stateString(profile, "otp_channel") == "" {
		profile["otp_channel"] = otpChannel
		profile["updated_at_unix"] = int64Now()
		_, _ = s.store.Save(ctx, gopayAccountProfileKey(key), stateJSON(profile))
	}
	return resolvedPaymentAccount{accountID: key, phone: phone, country: country, otpChannel: otpChannel, otpTarget: firstNonEmpty(phone, key), pin: stateString(profile, "pin"), state: state}, nil
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
	balanceCurrency := strings.ToUpper(firstNonEmpty(anyString(check["balance_currency"]), "IDR"))
	if currency != "" && balanceCurrency != "" && !strings.EqualFold(balanceCurrency, currency) {
		return fmt.Errorf("gopay balance currency mismatch: %s != %s", balanceCurrency, currency)
	}
	if balance < amount {
		return fmt.Errorf("insufficient gopay balance: %d %s < %d %s", balance, balanceCurrency, amount, firstNonEmpty(currency, balanceCurrency))
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
