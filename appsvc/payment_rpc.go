package appsvc

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
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
	currency := strings.ToUpper(stringx.FirstNonEmpty(req.GetCurrency(), stateString(account.state, "balance_currency"), "IDR"))
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
