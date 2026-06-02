package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func (h gopayHTTPHandler) handlePaymentAction(w http.ResponseWriter, r *http.Request, scope string, action string) {
	h.handleAction(w, r, scope, action, func(ctx context.Context, action string, req gopayActionRequest) (*gopayActionResult, error) {
		return h.invokeGoPayPaymentAction(ctx, scope, action, req)
	})
}

func (h gopayHTTPHandler) handleAccountAction(w http.ResponseWriter, r *http.Request, action string) {
	h.handleAction(w, r, gopayAccountActionScope, action, h.invokeGoPayAccountAction)
}

func (h gopayHTTPHandler) handleToolboxAction(w http.ResponseWriter, r *http.Request, action string) {
	h.handleAction(w, r, gopayToolboxActionScope, action, h.invokeGoPayToolboxAction)
}

func (h gopayHTTPHandler) handleAction(w http.ResponseWriter, r *http.Request, scope string, action string, invoker gopayActionInvoker) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	req, err := readGopayActionRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if invoker == nil {
		writeError(w, http.StatusBadGateway, fmt.Errorf("gopay action invoker is required"))
		return
	}
	req.actionScope = strings.TrimSpace(scope)
	result, err := invoker(r.Context(), action, req)
	if err != nil {
		if result != nil {
			result.Success = false
			result.ErrorMessage = err.Error()
			writeActionJSON(w, http.StatusOK, result)
			return
		}
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeActionJSON(w, http.StatusOK, result)
}
