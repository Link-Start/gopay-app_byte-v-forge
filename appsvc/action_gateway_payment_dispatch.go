package appsvc

import (
	"context"
	"fmt"
	"strings"
)

func (h gopayHTTPHandler) invokeGoPayPaymentAction(ctx context.Context, scope string, action string, req gopayActionRequest) (*gopayActionResult, error) {
	if isGoPayPaymentAccountRuntimeAction(action) {
		req.actionScope = scope
		req.Operation = normalizeGoPayWorkflowOperation(req.Operation)
		req.OTPChannel = normalizeActionOTPChannel(req.OTPChannel)
		return h.invokeGoPayAccountRuntimeAction(ctx, action, req, nil)
	}
	switch strings.TrimSpace(action) {
	case "start-payment":
		return h.startPayment(ctx, scope, req)
	case "complete-payment":
		return h.completePayment(ctx, scope, req)
	case "confirm-manual-payment":
		return h.confirmPayment(ctx, scope, req)
	case "cancel-payment", "cancel":
		return h.cancelPayment(ctx, scope, req)
	case "request-payment-otp", "resend-payment-otp", "resend-otp":
		return h.resendPaymentOTP(ctx, scope, req)
	case "await-channel-otp":
		return h.awaitPaymentChannelOTP(ctx, scope, req)
	default:
		return nil, fmt.Errorf("unsupported gopay payment action: %s", action)
	}
}

func isGoPayPaymentAccountRuntimeAction(action string) bool {
	switch strings.TrimSpace(action) {
	case "load-state",
		"start-gopay-auth",
		"complete-gopay-auth",
		"start-pin",
		"start-gopay-pin",
		"request-pin-otp",
		"retry-pin",
		"retry-gopay-pin",
		"complete-pin",
		"complete-gopay-pin",
		"acquire-signup-phone",
		"generate-signup-device-proxy",
		"check-signup-phone",
		"discard-signup-phone",
		"start-signup",
		"request-signup-otp",
		"retry-signup",
		"complete-signup",
		"acquire-change-phone-number",
		"start-change-phone",
		"retry-change-phone-otp",
		"complete-change-phone",
		"cancel-change-phone",
		"await-channel-otp":
		return true
	default:
		return false
	}
}
