package appsvc

import (
	"fmt"
	"net/http"

	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) handleWorkflowStart(w http.ResponseWriter, r *http.Request, key string) {
	switch key {
	case gopayAccountActionScope:
		h.handleAccountWorkflowStart(w, r)
	case gopayRegisterIndonesiaWAWorkflowKey:
		h.handleRegisterIndonesiaWAWorkflowStart(w, r)
	default:
		writeError(w, http.StatusNotFound, fmt.Errorf("unknown gopay workflow: %s", key))
	}
}

func (h gopayHTTPHandler) handleAccountWorkflowStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req pb.GoPayAccountWorkflowRequest
	if err := protojsonhttp.ReadRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	job, err := newGoPayWorkflowJob(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.saveWorkflowJob(r.Context(), job); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	h.triggerGoPayWorkflowAsync(job, gopayAccountWebhookPath, "gopay-account")
	_ = protojsonhttp.WriteResponse(w, http.StatusAccepted, &pb.GoPayAccountWorkflowResponse{JobId: job.GetJobId(), Started: true})
}

func (h gopayHTTPHandler) handleRegisterIndonesiaWAWorkflowStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	job := newGoPayRegisterIndonesiaWAWorkflowJob()
	if err := h.saveWorkflowJob(r.Context(), job); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	h.triggerGoPayWorkflowAsync(job, gopayRegisterIndonesiaWAWebhookPath, gopayRegisterIndonesiaWAWorkflowKey)
	_ = protojsonhttp.WriteResponse(w, http.StatusAccepted, &pb.StartGoPayRegisterIndonesiaWAWorkflowResponse{JobId: job.GetJobId(), Started: true})
}
