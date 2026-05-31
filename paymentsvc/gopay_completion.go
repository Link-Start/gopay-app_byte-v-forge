package paymentsvc

import (
	"context"
	"fmt"
	"strings"
)

func (c *charger) completeAfterOTPUntilManualConfirmation(ctx context.Context, state map[string]any, otp string) (map[string]any, error) {
	referenceID := stringAt(state, "reference_id")
	snapToken := stringAt(state, "snap_token")
	if referenceID == "" || snapToken == "" {
		return nil, fmt.Errorf("payment flow state is missing reference_id/snap_token")
	}
	if strings.TrimSpace(otp) == "" {
		return nil, fmt.Errorf("OTP not provided")
	}
	challengeID, clientID, err := c.gopayValidateOTP(ctx, referenceID, otp)
	if err != nil {
		return nil, err
	}
	pinToken, err := c.tokenizePIN(ctx, challengeID, clientID, true)
	if err != nil {
		return nil, err
	}
	if err := c.gopayValidatePIN(ctx, referenceID, pinToken); err != nil {
		return nil, err
	}
	chargeData, err := c.midtransCreateChargeData(ctx, snapToken)
	if err != nil {
		return nil, err
	}
	next := cloneMap(state)
	for key, value := range chargeData {
		next[key] = value
	}
	next["state"] = "awaiting_manual_confirmation"
	return next, nil
}

func (c *charger) completeAfterManualConfirmation(ctx context.Context, state map[string]any) (map[string]any, error) {
	snapToken := stringAt(state, "snap_token")
	csID := stringAt(state, "cs_id")
	chargeRef := stringAt(state, "charge_ref")
	if chargeRef == "" || snapToken == "" {
		return nil, fmt.Errorf("payment flow state is missing charge_ref/snap_token")
	}
	var midtransStatus map[string]any
	var err error
	if c.requiresManualConfirmation() {
		midtransStatus, err = c.midtransPollStatus(ctx, snapToken)
		if err != nil {
			return nil, err
		}
		c.followMidtransFinishRedirect(ctx, state, midtransStatus)
	} else {
		if err := c.gopayPaymentValidate(ctx, chargeRef); err != nil {
			return nil, err
		}
		challengeID, clientID, err := c.gopayPaymentConfirm(ctx, chargeRef)
		if err != nil {
			return nil, err
		}
		pinToken, err := c.tokenizePIN(ctx, challengeID, clientID, false)
		if err != nil {
			return nil, err
		}
		if err := c.gopayPaymentProcess(ctx, chargeRef, pinToken); err != nil {
			return nil, err
		}
		midtransStatus, err = c.midtransPollStatus(ctx, snapToken)
		if err != nil {
			return nil, err
		}
	}
	result := map[string]any{"state": "succeeded", "snap_token": snapToken, "charge_ref": chargeRef, "midtrans_status": stringAt(midtransStatus, "transaction_status")}
	if csID != "" {
		result["cs_id"] = csID
	}
	for _, key := range []string{"deeplink_url", "qr_code_url", "qr_string", "finish_redirect_url", "finish_200_redirect_url"} {
		result[key] = stringAt(state, key)
	}
	return result, nil
}

func (c *charger) completeAfterOTP(ctx context.Context, state map[string]any, otp string) (map[string]any, error) {
	next, err := c.completeAfterOTPUntilManualConfirmation(ctx, state, otp)
	if err != nil {
		return nil, err
	}
	return c.completeAfterManualConfirmation(ctx, next)
}
