package appsvc

import (
	"context"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) loadParamsResult(job *pb.GopayWorkflowJob, req gopayActionRequest) *gopayActionResult {
	out := h.baseResult(req, "load_params", true, map[string]any{"action": stringx.FirstNonEmpty(req.actionScope, gopayAccountActionScope), "operation": req.Operation, "gopay_account_id": req.GopayAccountID, "otp_channel": req.OTPChannel})
	out.RequirePIN = req.Operation == "ensure_pin_setup"
	return out
}

func (h gopayHTTPHandler) loadState(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.LoadGopayAccountState(ctx, &pb.LoadGopayAccountStateRequest{GopayAccountId: req.GopayAccountID})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "load_gopay_state", resp.GetSuccess(), nil)
	out.StateJSON = stringx.FirstNonEmpty(resp.GetStateJson(), "{}")
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) checkBalance(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	state := h.service.parseRequestState(stringx.FirstNonEmpty(req.StateJSON, "{}"))
	result := h.service.checkBalance(ctx, state)
	success := anyBool(result["success"])
	out := h.baseResult(req, "status", success, map[string]any{
		"balance_amount":   anyInt(result["balance_amount"]),
		"balance_currency": anyString(result["balance_currency"]),
		"has_min_balance":  anyBool(result["has_min_balance"]),
		"status":           anyInt(result["status"]),
	})
	out.StateJSON = stateJSON(state)
	out.AccountTokenReady = success
	out.Ready = success
	out.Phone = stateString(state, "phone")
	out.ErrorMessage = anyString(result["error"])
	_ = h.saveStateIfAccount(ctx, req, out.StateJSON)
	return out, nil
}

func (h gopayHTTPHandler) checkPIN(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.Status(ctx, &pb.StatusRequest{StateJson: stringx.FirstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	ready := resp.GetStage() == "ready"
	out := h.baseResult(req, "status", ready, map[string]any{
		"stage":            resp.GetStage(),
		"pin_setup":        resp.GetPinSetup(),
		"phone":            resp.GetPhone(),
		"balance_amount":   resp.GetBalanceAmount(),
		"balance_currency": resp.GetBalanceCurrency(),
	})
	out.StateJSON = resp.GetStateJson()
	out.AccountTokenReady = ready
	out.Ready = ready
	out.Phone = resp.GetPhone()
	out.SignupPINComplete = resp.GetPinSetup()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfAccount(ctx, req, out.StateJSON)
	return out, nil
}
