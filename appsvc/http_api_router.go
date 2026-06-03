package appsvc

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/gopay"), "/")
	switch {
	case path == "" || path == "health":
		h.handleHealth(w, r)
	case path == "action-catalog":
		h.handleActionCatalog(w, r)
	case path == "accounts":
		h.handleAccounts(w, r)
	case strings.HasPrefix(path, "accounts/"):
		h.handleAccount(w, r, strings.TrimPrefix(path, "accounts/"))
	case path == "profile":
		h.handleProfile(w, r)
	case path == "otp/submit":
		h.handleOTPSubmit(w, r)
	case path == "phone/check":
		h.handlePhoneCheck(w, r)
	case path == "settings":
		h.handleSettings(w, r)
	case strings.HasPrefix(path, "workflows/"):
		h.handleWorkflowStart(w, r, strings.Trim(strings.TrimPrefix(path, "workflows/"), "/"))
	case strings.HasPrefix(path, "actions/gopay-account/"):
		h.handleAccountAction(w, r, strings.Trim(strings.TrimPrefix(path, "actions/gopay-account/"), "/"))
	case strings.HasPrefix(path, "actions/gopay-toolbox/"):
		h.handleToolboxAction(w, r, strings.Trim(strings.TrimPrefix(path, "actions/gopay-toolbox/"), "/"))
	case strings.HasPrefix(path, "actions/gopay-payment/"):
		h.handlePaymentAction(w, r, gopayPaymentActionScope, strings.Trim(strings.TrimPrefix(path, "actions/gopay-payment/"), "/"))
	default:
		writeError(w, http.StatusNotFound, fmt.Errorf("unknown gopay API path: %s", path))
	}
}

func (h gopayHTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	workflows := make([]*pb.GopayHealthWorkflow, 0, len(goPayAccountWorkflows())+1)
	for _, workflow := range goPayAccountWorkflows() {
		workflows = append(workflows, &pb.GopayHealthWorkflow{Key: workflow.Key, Label: workflow.Label, WebhookPath: workflow.WebhookPath})
	}
	workflows = append(workflows, &pb.GopayHealthWorkflow{Key: gopayRegisterIndonesiaWAWorkflowKey, Label: registerIndonesiaWAWorkflowDisplayLabel, WebhookPath: gopayRegisterIndonesiaWAWebhookPath})
	_ = protojsonhttp.WriteResponse(w, http.StatusOK, &pb.GopayHealthResponse{
		Success:              true,
		Ok:                   true,
		Service:              "gopay-app",
		N8NWebhookConfigured: h.n8nWebhookBaseURL != "",
		GopayActionApiOwned:  true,
		GopayAccountApiOwned: true,
		Workflows:            workflows,
	})
}
