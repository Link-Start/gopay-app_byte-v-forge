package appsvc

import (
	"context"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) startDeactivate(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.DeactivateStart(ctx, &pb.DeactivateStartRequest{Pin: req.Pin, StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "deactivate_start", resp.GetSuccess(), nil)
	out.applyOTP(resp.GetOtpSent(), "sms")
	out.StateJSON = resp.GetStateJson()
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) completeDeactivate(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.DeactivateComplete(ctx, &pb.DeactivateCompleteRequest{Otp: req.otpValue(), StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "deactivate_complete", resp.GetSuccess(), map[string]any{"deactivated_at": resp.GetDeactivatedAt()})
	out.StateJSON = resp.GetStateJson()
	out.DeactivateComplete = resp.GetSuccess()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfAccount(ctx, req, out.StateJSON)
	return out, nil
}
