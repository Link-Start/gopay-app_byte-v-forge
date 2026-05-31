package paymentsvc

import (
	"context"
	"strings"

	"github.com/byte-v-forge/gopay-app/pb"
)

type StartInput struct {
	SnapToken         string
	CheckoutURL       string
	CheckoutSessionID string
	GopayAccountID    string
	Phone             string
	CountryCode       string
	PIN               string
	OTPChannel        string
	OTPTarget         string
	Tokenization      string
	Amount            int64
	Currency          string
}

func (s *Server) StartPayment(ctx context.Context, input StartInput) (*pb.StartGopayPaymentResponse, error) {
	snapToken := strings.TrimSpace(input.SnapToken)
	if snapToken == "" {
		return &pb.StartGopayPaymentResponse{Success: false, ErrorMessage: "snap_token is required"}, nil
	}
	ch, err := s.newCharger(ctx, input)
	if err != nil {
		return &pb.StartGopayPaymentResponse{Success: false, ErrorMessage: truncateError(err)}, nil
	}
	state := map[string]any{
		"state":               "prepared",
		"snap_token":          snapToken,
		"checkout_url":        strings.TrimSpace(input.CheckoutURL),
		"cs_id":               strings.TrimSpace(input.CheckoutSessionID),
		"checkout_session_id": strings.TrimSpace(input.CheckoutSessionID),
		"gopay_account_id":    strings.TrimSpace(input.GopayAccountID),
		"amount":              input.Amount,
		"currency":            strings.ToUpper(strings.TrimSpace(input.Currency)),
		"otp_channel":         normalizeOTPChannel(input.OTPChannel),
		"otp_target":          strings.TrimSpace(input.OTPTarget),
	}
	if checkoutURL := strings.TrimSpace(input.CheckoutURL); checkoutURL != "" {
		ch.checkoutURL = checkoutURL
	}
	var started map[string]any
	if ch.requiresManualConfirmation() {
		started, err = ch.startPreparedQRISToPaymentCharge(ctx, state)
	} else {
		started, err = ch.startPreparedLinkingUntilOTP(ctx, state, input.OTPChannel)
	}
	if err != nil {
		ch.close()
		return &pb.StartGopayPaymentResponse{Success: false, ErrorMessage: truncateError(err)}, nil
	}
	flowID := s.flows.put(&pendingFlow{charger: ch, state: started})
	return startResponse(flowID, started), nil
}

func (s *Server) ResendGopayPaymentOTP(ctx context.Context, req *pb.ResendGopayPaymentOTPRequest) (*pb.ResendGopayPaymentOTPResponse, error) {
	flowID := strings.TrimSpace(req.GetFlowId())
	if flowID == "" {
		return &pb.ResendGopayPaymentOTPResponse{Success: false, ErrorMessage: "flow_id is required"}, nil
	}
	flow := s.flows.get(flowID)
	if flow == nil {
		return &pb.ResendGopayPaymentOTPResponse{Success: false, ErrorMessage: "payment flow not found"}, nil
	}
	state, err := flow.charger.resendLinkingOTP(ctx, flow.state)
	if err != nil {
		return &pb.ResendGopayPaymentOTPResponse{Success: false, FlowId: flowID, ErrorMessage: truncateError(err)}, nil
	}
	flow.state = state
	return &pb.ResendGopayPaymentOTPResponse{Success: true, FlowId: flowID, IssuedAfterUnix: int64(intAt(state, "issued_after_unix"))}, nil
}

func (s *Server) CompleteGopayPayment(ctx context.Context, req *pb.CompleteGopayPaymentRequest) (*pb.GopayPaymentResponse, error) {
	flowID := strings.TrimSpace(req.GetFlowId())
	if flowID == "" {
		return &pb.GopayPaymentResponse{Success: false, ErrorMessage: "flow_id is required"}, nil
	}
	flow := s.flows.get(flowID)
	if flow == nil {
		return &pb.GopayPaymentResponse{Success: false, ErrorMessage: "payment flow not found"}, nil
	}
	if boolAt(flow.state, "otp_required") && strings.TrimSpace(req.GetOtp()) == "" {
		return &pb.GopayPaymentResponse{Success: false, ErrorMessage: "otp is required"}, nil
	}
	closeFlow := true
	var result map[string]any
	var err error
	if flow.charger.requiresManualConfirmation() {
		if boolAt(flow.state, "otp_required") {
			result, err = flow.charger.completeAfterOTPUntilManualConfirmation(ctx, flow.state, req.GetOtp())
			flow.state = result
		} else {
			result = flow.state
		}
		closeFlow = false
	} else if boolAt(flow.state, "otp_required") {
		result, err = flow.charger.completeAfterOTP(ctx, flow.state, req.GetOtp())
	} else {
		result, err = flow.charger.completeAfterManualConfirmation(ctx, flow.state)
	}
	if err != nil {
		if closeFlow {
			if failed := s.flows.pop(flowID); failed != nil {
				failed.close()
			}
		}
		return &pb.GopayPaymentResponse{Success: false, ErrorMessage: truncateError(err)}, nil
	}
	if closeFlow {
		if done := s.flows.pop(flowID); done != nil {
			done.close()
		}
	} else {
		flow.state = result
	}
	return paymentResponse(result, !closeFlow), nil
}

func (s *Server) ConfirmGopayPayment(ctx context.Context, req *pb.ConfirmGopayPaymentRequest) (*pb.GopayPaymentResponse, error) {
	flowID := strings.TrimSpace(req.GetFlowId())
	if flowID == "" {
		return &pb.GopayPaymentResponse{Success: false, ErrorMessage: "flow_id is required"}, nil
	}
	flow := s.flows.get(flowID)
	if flow == nil {
		return &pb.GopayPaymentResponse{Success: false, ErrorMessage: "payment flow not found"}, nil
	}
	result, err := flow.charger.completeAfterManualConfirmation(ctx, flow.state)
	if err != nil {
		return &pb.GopayPaymentResponse{Success: false, ErrorMessage: truncateError(err)}, nil
	}
	if done := s.flows.pop(flowID); done != nil {
		done.close()
	}
	return paymentResponse(result, false), nil
}

func (s *Server) CancelGopayPayment(_ context.Context, req *pb.CancelGopayPaymentRequest) (*pb.CancelGopayPaymentResponse, error) {
	if req == nil {
		return &pb.CancelGopayPaymentResponse{Success: true}, nil
	}
	if flowID := strings.TrimSpace(req.GetFlowId()); flowID != "" {
		if flow := s.flows.pop(flowID); flow != nil {
			flow.close()
		}
	}
	return &pb.CancelGopayPaymentResponse{Success: true}, nil
}
