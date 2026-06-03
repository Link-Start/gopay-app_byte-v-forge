package appsvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/protojsonx"
	"github.com/byte-v-forge/gopay-app/pb"
	"github.com/google/uuid"
)

func newGoPayWorkflowJob(req *pb.GoPayAccountWorkflowRequest) (*pb.GopayWorkflowJob, error) {
	return newGoPayWorkflowJobWithOperation(req, "")
}

func newGoPayWorkflowJobWithOperation(req *pb.GoPayAccountWorkflowRequest, operationOverride string) (*pb.GopayWorkflowJob, error) {
	operation := normalizeGoPayWorkflowOperation(operationOverride)
	if operation == "" || operation == "login" && operationOverride == "" {
		operation = normalizeGoPayWorkflowOperation(req.GetOperation().String())
	}
	accountID := goPayAppAccountID(req.GetGopayAccountId())
	if accountID == "" {
		return nil, fmt.Errorf("gopay_account_id is required")
	}
	now := time.Now().Unix()
	return &pb.GopayWorkflowJob{
		JobId:          uuid.NewString(),
		Operation:      operation,
		GopayAccountId: accountID,
		Phone:          strings.TrimSpace(req.GetPhone()),
		CountryCode:    strings.TrimSpace(req.GetCountryCode()),
		Pin:            strings.TrimSpace(req.GetPin()),
		OtpChannel:     normalizeActionOTPChannel(req.GetOtpChannel()),
		Status:         "started",
		CreatedAtUnix:  now,
		UpdatedAtUnix:  now,
	}, nil
}

func newGoPayRegisterIndonesiaWAWorkflowJob() *pb.GopayWorkflowJob {
	now := time.Now().Unix()
	return &pb.GopayWorkflowJob{
		JobId:         uuid.NewString(),
		Operation:     registerIndonesiaWAWorkflowOperation,
		Status:        "started",
		CreatedAtUnix: now,
		UpdatedAtUnix: now,
	}
}

func (h gopayHTTPHandler) triggerGoPayWorkflow(ctx context.Context, jobID string, webhookPath string, workflowName string) error {
	if h.n8nWebhookBaseURL == "" {
		return fmt.Errorf("GOPAY_N8N_WEBHOOK_BASE_URL is required")
	}
	url := h.n8nWebhookBaseURL + "/" + strings.Trim(strings.TrimSpace(webhookPath), "/")
	return postJSON(ctx, url, map[string]string{"job_id": strings.TrimSpace(jobID)}, jsonPostOptions{
		Timeout:   30 * time.Second,
		Operation: "n8n " + strings.TrimSpace(workflowName) + " workflow",
	})
}

func (h gopayHTTPHandler) triggerGoPayWorkflowAsync(job *pb.GopayWorkflowJob, webhookPath string, workflowName string) {
	if job == nil {
		return
	}
	go func(snapshot pb.GopayWorkflowJob) {
		if err := h.triggerGoPayWorkflow(context.Background(), snapshot.GetJobId(), webhookPath, workflowName); err != nil {
			snapshot.Status = "trigger_failed"
			snapshot.ErrorMessage = err.Error()
			snapshot.UpdatedAtUnix = time.Now().Unix()
			_ = h.saveWorkflowJob(context.Background(), &snapshot)
		}
	}(*job)
}

func (h gopayHTTPHandler) loadWorkflowJob(ctx context.Context, jobID string) (*pb.GopayWorkflowJob, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	raw, err := h.service.store.Load(ctx, workflowJobKey(jobID))
	if err != nil {
		return nil, err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = "{}"
	}
	var job pb.GopayWorkflowJob
	if err := protojsonx.Unmarshal([]byte(raw), &job); err != nil {
		return nil, err
	}
	if strings.TrimSpace(job.GetJobId()) == "" {
		return nil, fmt.Errorf("gopay workflow job not found: %s", jobID)
	}
	return &job, nil
}

func (h gopayHTTPHandler) saveWorkflowJob(ctx context.Context, job *pb.GopayWorkflowJob) error {
	if job == nil || strings.TrimSpace(job.GetJobId()) == "" {
		return fmt.Errorf("job_id is required")
	}
	data, err := protojsonx.Marshal(job)
	if err != nil {
		return err
	}
	_, err = h.service.store.Save(ctx, workflowJobKey(job.GetJobId()), string(data))
	return err
}

func workflowJobKey(jobID string) string {
	return "workflow-job:" + strings.TrimSpace(jobID)
}
