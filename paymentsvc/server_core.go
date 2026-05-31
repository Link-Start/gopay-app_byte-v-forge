package paymentsvc

import "github.com/byte-v-forge/gopay-app/pb"

type Server struct {
	cfg   Config
	flows *flowStore
}

func NewServer(cfg Config) *Server {
	return &Server{cfg: cfg, flows: &flowStore{items: map[string]*pendingFlow{}}}
}

type pendingFlow struct {
	charger *charger
	state   map[string]any
}

func (f *pendingFlow) close() {
	if f != nil && f.charger != nil {
		f.charger.close()
	}
}

func startResponse(flowID string, state map[string]any) *pb.StartGopayPaymentResponse {
	return &pb.StartGopayPaymentResponse{
		Success:           true,
		FlowId:            flowID,
		SnapToken:         stringAt(state, "snap_token"),
		IssuedAfterUnix:   intAt(state, "issued_after_unix"),
		CheckoutUrl:       stringAt(state, "checkout_url"),
		CheckoutSessionId: firstNonEmpty(stringAt(state, "cs_id"), stringAt(state, "checkout_session_id")),
		OtpRequired:       boolAt(state, "otp_required"),
		OtpChannel:        stringAt(state, "otp_channel"),
		GopayAccountId:    stringAt(state, "gopay_account_id"),
		Amount:            intAt(state, "amount"),
		Currency:          stringAt(state, "currency"),
		OtpTarget:         stringAt(state, "otp_target"),
	}
}

func paymentResponse(result map[string]any, awaitingManual bool) *pb.GopayPaymentResponse {
	state := stringAt(result, "state")
	success := awaitingManual || state == "succeeded"
	message := ""
	if !success {
		message = "payment state=" + firstNonEmpty(state, "unknown")
	}
	return &pb.GopayPaymentResponse{
		Success:                    success,
		ErrorMessage:               message,
		ChargeRef:                  stringAt(result, "charge_ref"),
		SnapToken:                  stringAt(result, "snap_token"),
		AwaitingManualConfirmation: awaitingManual,
		DeeplinkUrl:                stringAt(result, "deeplink_url"),
		QrCodeUrl:                  stringAt(result, "qr_code_url"),
		FinishRedirectUrl:          stringAt(result, "finish_redirect_url"),
		Finish_200RedirectUrl:      stringAt(result, "finish_200_redirect_url"),
		QrString:                   stringAt(result, "qr_string"),
	}
}
