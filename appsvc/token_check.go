package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (s *Server) checkTokenValid(ctx context.Context, state stateMap) map[string]any {
	profile := s.verifyAccessToken(ctx, state)
	if anyBool(profile["success"]) {
		return s.tokenValidResult(ctx, state, profile, false)
	}
	refresh := s.refreshAccessToken(ctx, state)
	if !anyBool(refresh["success"]) {
		return map[string]any{"success": false, "token_valid": false, "refreshed": false, "error": stringx.FirstNonEmpty(anyString(refresh["error"]), anyString(profile["error"]), "token invalid")}
	}
	profile = s.verifyAccessToken(ctx, state)
	if anyBool(profile["success"]) {
		return s.tokenValidResult(ctx, state, profile, true)
	}
	return map[string]any{"success": false, "token_valid": false, "refreshed": true, "error": stringx.FirstNonEmpty(anyString(profile["error"]), "profile failed after refresh")}
}

func (s *Server) tokenValidResult(ctx context.Context, state stateMap, profile map[string]any, refreshed bool) map[string]any {
	balance := s.checkBalance(ctx, state)
	balanceOK := anyBool(balance["success"])
	amount := anyInt(balance["balance_amount"])
	currency := anyString(balance["balance_currency"])
	if !balanceOK {
		amount = firstNonZero(amount, stateInt(state, "balance_amount"))
		currency = stringx.FirstNonEmpty(currency, stateString(state, "balance_currency"))
	}
	cachedMinBalance := !balanceOK && (anyBool(state["has_min_balance"]) || stateInt(state, "balance_amount") >= s.cfg.MinBalanceRp)
	hasMinBalance := anyBool(balance["has_min_balance"]) || cachedMinBalance
	result := map[string]any{
		"success":          balanceOK || cachedMinBalance,
		"token_valid":      true,
		"refreshed":        refreshed,
		"phone":            profile["phone"],
		"balance_amount":   amount,
		"balance_currency": stringx.FirstNonEmpty(currency, "IDR"),
		"has_min_balance":  hasMinBalance,
	}
	if cachedMinBalance {
		result["cached_balance"] = true
		result["balance_check_error"] = stringx.FirstNonEmpty(anyString(balance["error"]), "balance check failed")
	}
	if !balanceOK && !cachedMinBalance {
		result["error"] = stringx.FirstNonEmpty(anyString(balance["error"]), "balance check failed")
	}
	return result
}

func (s *Server) checkBalance(ctx context.Context, state stateMap) map[string]any {
	if stateString(state, "token") == "" {
		return map[string]any{"success": false, "error": "access_token missing", "status": 0}
	}
	deleteKeys(state, "last_token_refresh_error", "last_token_refresh_failed_at")
	client, err := s.newClientWithState(ctx, state, false)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error(), "status": 0}
	}
	resp, err := client.Customer.Get(ctx, "/v1/payment-options/balances")
	state["last_balance_check_at"] = time.Now().Unix()
	if err != nil {
		state["last_balance_error"] = err.Error()
		return map[string]any{"success": false, "status": 0, "error": err.Error()}
	}
	if resp.StatusCode != http.StatusOK {
		errMessage := apiError("balance check failed", resp)
		state["last_balance_error"] = errMessage
		return map[string]any{"success": false, "status": resp.StatusCode, "error": errMessage}
	}
	if resp.Payload["success"] == false {
		errMessage := apiError("balance check failed", resp)
		state["last_balance_error"] = errMessage
		return map[string]any{"success": false, "status": resp.StatusCode, "error": errMessage}
	}
	amount, currency := walletBalance(resp.Payload["data"])
	if amount < 0 {
		errMessage := "gopay wallet balance missing"
		state["last_balance_error"] = errMessage
		return map[string]any{"success": false, "status": resp.StatusCode, "error": errMessage}
	}
	hasMin := amount >= s.cfg.MinBalanceRp
	state["balance_amount"] = amount
	state["balance_currency"] = stringx.FirstNonEmpty(currency, "IDR")
	state["has_min_balance"] = hasMin
	delete(state, "last_balance_error")
	if hasMin {
		if stateString(state, "last_error") == "INSUFFICIENT_GOPAY_BALANCE" {
			delete(state, "last_error")
		}
	} else {
		state["last_error"] = "INSUFFICIENT_GOPAY_BALANCE"
	}
	return map[string]any{"success": true, "status": 200, "balance_amount": amount, "balance_currency": stateString(state, "balance_currency"), "has_min_balance": hasMin}
}

func (s *Server) tokenCheckReady(result map[string]any) bool {
	return anyBool(result["success"]) && anyBool(result["token_valid"]) && anyBool(result["has_min_balance"])
}

func (s *Server) tokenCheckValid(result map[string]any) bool {
	return anyBool(result["token_valid"])
}

func (s *Server) tokenCheckError(result map[string]any) string {
	if err := anyString(result["error"]); err != "" {
		return err
	}
	amount := anyInt(result["balance_amount"])
	currency := stringx.FirstNonEmpty(anyString(result["balance_currency"]), "IDR")
	return fmt.Sprintf("insufficient gopay balance: %d %s < %d IDR", amount, currency, s.cfg.MinBalanceRp)
}
