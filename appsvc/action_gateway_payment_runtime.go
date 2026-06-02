package appsvc

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) startPayment(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	tokenization := stringx.FirstNonEmpty(req.Tokenization, "true")
	resp, err := h.service.StartGopayPayment(ctx, &pb.StartGopayPaymentRequest{
		SnapToken:         req.SnapToken,
		CheckoutUrl:       req.CheckoutURL,
		CheckoutSessionId: req.CheckoutSessionID,
		Tokenization:      tokenization,
		GopayAccountId:    req.GopayAccountID,
		Amount:            req.Amount,
		Currency:          req.Currency,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		resp = &pb.StartGopayPaymentResponse{Success: false, ErrorMessage: "empty gopay payment start response"}
	}
	data := paymentResponseData(resp, nil)
	data["runtime_owner"] = "gopay-app"
	out := h.basePaymentResult(scope, req, "gopay_payment", resp.GetSuccess(), data)
	out.GopayAccountID = resp.GetGopayAccountId()
	out.OTPChannel = resp.GetOtpChannel()
	out.FlowID = resp.GetFlowId()
	out.CheckoutURL = resp.GetCheckoutUrl()
	out.CheckoutSessionID = resp.GetCheckoutSessionId()
	out.UseAccountToken = req.UseAccountToken
	out.OTPRequired = resp.GetOtpRequired()
	out.OTPIssuedAfterUnix = resp.GetIssuedAfterUnix()
	out.OTPTimeoutSeconds = firstNonZeroInt32(req.OTPTimeoutSeconds, 300)
	out.SnapTokenPresent = resp.GetSnapToken() != ""
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) completePayment(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CompleteGopayPayment(ctx, &pb.CompleteGopayPaymentRequest{
		FlowId: req.FlowID,
		Otp:    req.otpValue(),
	})
	if err != nil {
		return nil, err
	}
	return h.paymentResult(scope, req, "gopay_payment", resp), nil
}

func (h gopayHTTPHandler) confirmPayment(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.ConfirmGopayPayment(ctx, &pb.ConfirmGopayPaymentRequest{FlowId: req.FlowID})
	if err != nil {
		return nil, err
	}
	return h.paymentResult(scope, req, "gopay_payment", resp), nil
}

func (h gopayHTTPHandler) cancelPayment(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CancelGopayPayment(ctx, &pb.CancelGopayPaymentRequest{FlowId: req.FlowID})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		resp = &pb.CancelGopayPaymentResponse{Success: false, ErrorMessage: "empty gopay payment cancel response"}
	}
	data := map[string]any{"flow_id": strings.TrimSpace(req.FlowID), "runtime_owner": "gopay-app"}
	out := h.basePaymentResult(scope, req, "cancel_payment", resp.GetSuccess(), data)
	out.FlowID = strings.TrimSpace(req.FlowID)
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) resendPaymentOTP(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.ResendGopayPaymentOTP(ctx, &pb.ResendGopayPaymentOTPRequest{FlowId: req.FlowID})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		resp = &pb.ResendGopayPaymentOTPResponse{Success: false, ErrorMessage: "empty gopay payment otp response"}
	}
	data := map[string]any{
		"response_present":      true,
		"success":               resp.GetSuccess(),
		"error_message":         resp.GetErrorMessage(),
		"flow_id":               resp.GetFlowId(),
		"otp_issued_after_unix": resp.GetIssuedAfterUnix(),
		"runtime_owner":         "gopay-app",
	}
	out := h.basePaymentResult(scope, req, "resend_payment_otp", resp.GetSuccess(), data)
	out.FlowID = resp.GetFlowId()
	out.OTPIssuedAfterUnix = resp.GetIssuedAfterUnix()
	out.OTPRequired = true
	out.OTPTimeoutSeconds = firstNonZeroInt32(req.OTPTimeoutSeconds, 300)
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) paymentResult(scope string, req gopayActionRequest, step string, resp *pb.GopayPaymentResponse) *gopayActionResult {
	if resp == nil {
		resp = &pb.GopayPaymentResponse{Success: false, ErrorMessage: "empty gopay payment response"}
	}
	data := paymentResponseData(nil, resp)
	data["runtime_owner"] = "gopay-app"
	out := h.basePaymentResult(scope, req, step, resp.GetSuccess(), data)
	out.FlowID = req.FlowID
	out.ChargeRef = resp.GetChargeRef()
	out.SnapTokenPresent = resp.GetSnapToken() != ""
	out.AwaitingManual = resp.GetAwaitingManualConfirmation()
	settled := resp.GetSuccess() && !resp.GetAwaitingManualConfirmation() && resp.GetChargeRef() != ""
	out.PlusTrialEligible = settled
	out.PlusTrialChecked = settled
	out.PlusActive = settled
	out.ErrorMessage = resp.GetErrorMessage()
	return out
}

func paymentResponseData(start *pb.StartGopayPaymentResponse, result *pb.GopayPaymentResponse) map[string]any {
	data := map[string]any{"response_present": start != nil || result != nil}
	if start != nil {
		data["success"] = start.GetSuccess()
		data["error_message"] = start.GetErrorMessage()
		data["flow_id"] = start.GetFlowId()
		data["snap_token_present"] = start.GetSnapToken() != ""
		data["issued_after_unix"] = start.GetIssuedAfterUnix()
		data["expires_at_unix"] = start.GetExpiresAtUnix()
		data["checkout_url"] = start.GetCheckoutUrl()
		data["checkout_session_id"] = start.GetCheckoutSessionId()
		data["otp_required"] = start.GetOtpRequired()
		data["otp_channel"] = start.GetOtpChannel()
		data["gopay_account_id"] = start.GetGopayAccountId()
		data["amount"] = start.GetAmount()
		data["currency"] = start.GetCurrency()
		data["otp_target"] = start.GetOtpTarget()
	}
	if result != nil {
		data["success"] = result.GetSuccess()
		data["error_message"] = result.GetErrorMessage()
		data["charge_ref"] = result.GetChargeRef()
		data["snap_token_present"] = result.GetSnapToken() != ""
		data["awaiting_manual_confirmation"] = result.GetAwaitingManualConfirmation()
		data["deeplink_url"] = result.GetDeeplinkUrl()
		data["qr_code_url"] = result.GetQrCodeUrl()
		data["qr_string"] = result.GetQrString()
		data["finish_redirect_url"] = result.GetFinishRedirectUrl()
		data["finish_200_redirect_url"] = result.GetFinish_200RedirectUrl()
	}
	return data
}
