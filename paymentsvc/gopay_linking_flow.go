package paymentsvc

import (
	"context"
	"fmt"
	"time"
)

func (c *charger) startLinkingUntilOTP(ctx context.Context, snapToken, csID, stripePK, otpChannel string) (map[string]any, error) {
	otpChannel = normalizeOTPChannel(otpChannel)
	if err := c.midtransLoadTransaction(ctx, snapToken); err != nil {
		return nil, err
	}
	referenceID, err := c.midtransInitLinking(ctx, snapToken)
	if err != nil {
		return nil, err
	}
	if err := c.gopayValidateReference(ctx, referenceID); err != nil {
		return nil, err
	}
	issued := time.Now().Unix()
	consent, err := c.gopayLinkingConsent(ctx, referenceID)
	if err != nil {
		return nil, err
	}
	otpRequired := linkingConsentRequiresOTP(consent)
	if otpRequired && goPayOTPChannelRequiresSMSActivation(otpChannel) {
		_, _ = c.gopayResendOTP(ctx, referenceID)
	}
	state := map[string]any{
		"cs_id":             csID,
		"checkout_url":      c.checkoutURL,
		"stripe_pk":         stripePK,
		"snap_token":        snapToken,
		"reference_id":      referenceID,
		"issued_after_unix": issued,
		"otp_channel":       otpChannel,
		"otp_required":      otpRequired,
	}
	if !otpRequired {
		chargeData, err := c.midtransCreateChargeData(ctx, snapToken)
		if err != nil {
			return nil, err
		}
		for key, value := range chargeData {
			state[key] = value
		}
		state["state"] = "awaiting_manual_confirmation"
	}
	return state, nil
}

func (c *charger) resendLinkingOTP(ctx context.Context, state map[string]any) (map[string]any, error) {
	referenceID := stringAt(state, "reference_id")
	if referenceID == "" {
		return nil, fmt.Errorf("prepared payment is missing reference_id")
	}
	if _, err := c.gopayResendOTP(ctx, referenceID); err != nil {
		return nil, err
	}
	next := cloneMap(state)
	next["issued_after_unix"] = time.Now().Unix()
	next["otp_required"] = true
	next["otp_resend_count"] = intAt(state, "otp_resend_count") + 1
	return next, nil
}
