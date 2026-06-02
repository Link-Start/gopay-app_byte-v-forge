package appsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/gopay-app/pb"
	"github.com/google/uuid"
)

const (
	gopayAccountActionScope                 = "gopay-account"
	gopayToolboxActionScope                 = "gopay-toolbox"
	gopayPaymentActionScope                 = "gopay-payment"
	gopayAccountWebhookPath                 = "gopay-app/account"
	gopayRegisterIndonesiaWAWorkflowKey     = "register-indonesia-wa"
	gopayRegisterIndonesiaWAWebhookPath     = "gopay-app/toolbox/register-indonesia-wa"
	registerIndonesiaWAWorkflowOperation    = "register_indonesia_wa"
	registerIndonesiaWAWorkflowDisplayLabel = "注册印尼 WA"
)

type gopayActionRequest struct {
	JobID             string         `json:"job_id"`
	N8NExecutionID    string         `json:"n8n_execution_id"`
	Operation         string         `json:"operation"`
	GopayAccountID    string         `json:"gopay_account_id"`
	Phone             string         `json:"phone"`
	WAPhone           string         `json:"wa_phone"`
	CountryCode       string         `json:"country_code"`
	Pin               string         `json:"pin"`
	OTPChannel        string         `json:"otp_channel"`
	ActivationID      string         `json:"activation_id"`
	StateJSON         string         `json:"state_json"`
	OTPIssuedAfter    int64          `json:"otp_issued_after_unix"`
	OTPTimeoutSeconds int32          `json:"otp_timeout_seconds"`
	OTPSource         string         `json:"otp_source"`
	OTP               string         `json:"otp"`
	Channel           string         `json:"channel"`
	Target            string         `json:"target"`
	ResumeURL         string         `json:"resume_url"`
	FlowID            string         `json:"flow_id"`
	SnapToken         string         `json:"snap_token"`
	CheckoutURL       string         `json:"checkout_url"`
	CheckoutSessionID string         `json:"checkout_session_id"`
	Tokenization      string         `json:"tokenization"`
	Amount            int64          `json:"amount"`
	Currency          string         `json:"currency"`
	UseAccountToken   bool           `json:"use_account_token"`
	FailureCount      int32          `json:"failure_count"`
	OTPRetryAttempt   int32          `json:"otp_retry_attempt"`
	Reason            string         `json:"reason"`
	ErrorMessage      string         `json:"error_message"`
	Data              map[string]any `json:"data"`
	actionScope       string
}

type gopayWorkflowJob struct {
	JobID          string         `json:"job_id"`
	N8NExecutionID string         `json:"n8n_execution_id,omitempty"`
	Operation      string         `json:"operation"`
	GopayAccountID string         `json:"gopay_account_id"`
	Phone          string         `json:"phone,omitempty"`
	CountryCode    string         `json:"country_code,omitempty"`
	Pin            string         `json:"pin,omitempty"`
	OTPChannel     string         `json:"otp_channel,omitempty"`
	Status         string         `json:"status"`
	ErrorMessage   string         `json:"error_message,omitempty"`
	CreatedAtUnix  int64          `json:"created_at_unix"`
	UpdatedAtUnix  int64          `json:"updated_at_unix"`
	Result         map[string]any `json:"result,omitempty"`
}

type gopayActionResult struct {
	JobID               string         `json:"job_id"`
	N8NExecutionID      string         `json:"n8n_execution_id,omitempty"`
	Action              string         `json:"action"`
	Step                string         `json:"step"`
	Operation           string         `json:"operation,omitempty"`
	Success             bool           `json:"success"`
	Started             bool           `json:"started,omitempty"`
	GopayAccountID      string         `json:"gopay_account_id,omitempty"`
	Phone               string         `json:"phone,omitempty"`
	OTPChannel          string         `json:"otp_channel,omitempty"`
	OTPRequired         bool           `json:"otp_required,omitempty"`
	OTPSent             bool           `json:"otp_sent,omitempty"`
	OTPFound            bool           `json:"otp_found,omitempty"`
	OTPSource           string         `json:"otp_source,omitempty"`
	OTPIssuedAfterUnix  int64          `json:"otp_issued_after_unix,omitempty"`
	OTPTimeoutSeconds   int32          `json:"otp_timeout_seconds,omitempty"`
	CountryCode         string         `json:"country_code,omitempty"`
	ActivationID        string         `json:"activation_id,omitempty"`
	FailureCount        int32          `json:"failure_count,omitempty"`
	ErrorMessage        string         `json:"error_message,omitempty"`
	RequirePIN          bool           `json:"require_pin,omitempty"`
	StateJSON           string         `json:"state_json,omitempty"`
	Ready               bool           `json:"ready,omitempty"`
	AccountTokenReady   bool           `json:"account_token_ready,omitempty"`
	PhoneAccepted       bool           `json:"phone_accepted,omitempty"`
	RetryableFailure    bool           `json:"retryable_failure,omitempty"`
	RotatableFailure    bool           `json:"rotatable_failure,omitempty"`
	ChangePhoneComplete bool           `json:"change_phone_complete,omitempty"`
	DeactivateComplete  bool           `json:"deactivate_complete,omitempty"`
	SignupComplete      bool           `json:"signup_complete,omitempty"`
	SignupPINComplete   bool           `json:"signup_pin_complete,omitempty"`
	FlowID              string         `json:"flow_id,omitempty"`
	CheckoutURL         string         `json:"checkout_url,omitempty"`
	CheckoutSessionID   string         `json:"checkout_session_id,omitempty"`
	UseAccountToken     bool           `json:"use_account_token,omitempty"`
	AwaitingManual      bool           `json:"awaiting_manual_confirmation,omitempty"`
	ChargeRef           string         `json:"charge_ref,omitempty"`
	SnapTokenPresent    bool           `json:"snap_token_present,omitempty"`
	PlusTrialEligible   bool           `json:"plus_trial_eligible,omitempty"`
	PlusTrialChecked    bool           `json:"plus_trial_checked,omitempty"`
	PlusActive          bool           `json:"plus_active,omitempty"`
	Data                map[string]any `json:"data,omitempty"`
}

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
	_ = protojsonhttp.WriteResponse(w, http.StatusAccepted, &pb.GoPayAccountWorkflowResponse{JobId: job.JobID, Started: true})
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
	_ = protojsonhttp.WriteResponse(w, http.StatusAccepted, &pb.StartGoPayRegisterIndonesiaWAWorkflowResponse{JobId: job.JobID, Started: true})
}

func (h gopayHTTPHandler) handlePaymentAction(w http.ResponseWriter, r *http.Request, scope string, action string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	req, err := readGopayActionRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := h.invokeGoPayPaymentAction(r.Context(), scope, action, req)
	if err != nil {
		if result != nil {
			result.Success = false
			result.ErrorMessage = err.Error()
			writeJSON(w, http.StatusOK, result)
			return
		}
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h gopayHTTPHandler) handleAccountAction(w http.ResponseWriter, r *http.Request, action string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	req, err := readGopayActionRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.actionScope = gopayAccountActionScope
	result, err := h.invokeGoPayAccountAction(r.Context(), action, req)
	if err != nil {
		if result != nil {
			result.Success = false
			result.ErrorMessage = err.Error()
			writeJSON(w, http.StatusOK, result)
			return
		}
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h gopayHTTPHandler) handleToolboxAction(w http.ResponseWriter, r *http.Request, action string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	req, err := readGopayActionRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.actionScope = gopayToolboxActionScope
	result, err := h.invokeGoPayToolboxAction(r.Context(), action, req)
	if err != nil {
		if result != nil {
			result.Success = false
			result.ErrorMessage = err.Error()
			writeJSON(w, http.StatusOK, result)
			return
		}
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h gopayHTTPHandler) invokeGoPayAccountAction(ctx context.Context, action string, req gopayActionRequest) (*gopayActionResult, error) {
	job, loadErr := h.loadWorkflowJob(ctx, req.JobID)
	if strings.TrimSpace(req.JobID) != "" && job == nil && strings.TrimSpace(action) != "fail" {
		return nil, loadErr
	}
	req = req.withJob(job)
	return h.invokeGoPayAccountRuntimeAction(ctx, action, req, job)
}

func (h gopayHTTPHandler) invokeGoPayAccountRuntimeAction(ctx context.Context, action string, req gopayActionRequest, job *gopayWorkflowJob) (*gopayActionResult, error) {
	switch strings.TrimSpace(action) {
	case "load-params":
		return h.loadParamsResult(job, req), nil
	case "register-indonesia-wa-settings":
		return h.registerIndonesiaWASettings(ctx, req)
	case "generate-shared-phone-check-proxy":
		return h.generateSharedPhoneCheckProxy(ctx, req)
	case "release-shared-phone-check-proxy":
		return h.releaseSharedPhoneCheckProxy(ctx, req)
	case "load-state":
		return h.loadState(ctx, req)
	case "check-balance":
		return h.checkBalance(ctx, req)
	case "check-pin":
		return h.checkPIN(ctx, req)
	case "start-gopay-auth":
		return h.startAuth(ctx, req)
	case "complete-gopay-auth":
		return h.completeAuth(ctx, req)
	case "start-pin", "start-gopay-pin":
		return h.startPIN(ctx, req)
	case "request-pin-otp", "retry-pin", "retry-gopay-pin":
		return h.retryPIN(ctx, req)
	case "complete-pin", "complete-gopay-pin":
		return h.completePIN(ctx, req)
	case "acquire-signup-phone":
		return h.acquireSignupPhone(req), nil
	case "generate-signup-device-proxy":
		return h.generateSharedPhoneCheckProxy(ctx, req)
	case "check-signup-phone":
		return h.checkSignupPhone(ctx, req)
	case "discard-signup-phone":
		return h.discardSignupPhone(req), nil
	case "start-signup":
		return h.startSignup(ctx, req)
	case "request-signup-otp", "retry-signup":
		return h.retrySignup(ctx, req)
	case "complete-signup":
		return h.completeSignup(ctx, req)
	case "start-deactivate":
		return h.startDeactivate(ctx, req)
	case "complete-deactivate":
		return h.completeDeactivate(ctx, req)
	case "finish-sms", "cancel-change-phone", "request-signup-sms":
		return h.genericSuccess(req, action, nil), nil
	case "acquire-change-phone-number":
		return h.acquireChangePhone(req), nil
	case "start-change-phone":
		return h.startChangePhone(ctx, req)
	case "retry-change-phone-otp":
		return h.retryChangePhone(ctx, req)
	case "complete-change-phone":
		return h.completeChangePhone(ctx, req)
	case "await-channel-otp":
		return h.awaitChannelOTP(ctx, req)
	case "finish":
		return h.finish(ctx, req)
	case "fail":
		return h.fail(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported gopay account action: %s", action)
	}
}

func (h gopayHTTPHandler) invokeGoPayToolboxAction(ctx context.Context, action string, req gopayActionRequest) (*gopayActionResult, error) {
	job, loadErr := h.loadWorkflowJob(ctx, req.JobID)
	if strings.TrimSpace(req.JobID) != "" && job == nil && strings.TrimSpace(action) != "fail" {
		return nil, loadErr
	}
	req = req.withJob(job)
	req.actionScope = gopayToolboxActionScope
	switch strings.TrimSpace(action) {
	case "load-params":
		return h.loadParamsResult(job, req), nil
	case "register-indonesia-wa-settings":
		return h.registerIndonesiaWASettings(ctx, req)
	case "generate-shared-phone-check-proxy":
		return h.generateSharedPhoneCheckProxy(ctx, req)
	case "release-shared-phone-check-proxy":
		return h.releaseSharedPhoneCheckProxy(ctx, req)
	case "finish":
		return h.finish(ctx, req)
	case "fail":
		return h.fail(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported gopay toolbox action: %s", action)
	}
}

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

func (h gopayHTTPHandler) loadParamsResult(job *gopayWorkflowJob, req gopayActionRequest) *gopayActionResult {
	out := h.baseResult(req, "load_params", true, map[string]any{"action": firstNonEmpty(req.actionScope, gopayAccountActionScope), "operation": req.Operation, "gopay_account_id": req.GopayAccountID, "otp_channel": req.OTPChannel})
	out.RequirePIN = req.Operation == "provision"
	return out
}

func (h gopayHTTPHandler) loadState(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.LoadGopayAccountState(ctx, &pb.LoadGopayAccountStateRequest{GopayAccountId: req.GopayAccountID})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "load_gopay_state", resp.GetSuccess(), nil)
	out.StateJSON = firstNonEmpty(resp.GetStateJson(), "{}")
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) checkBalance(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	state := h.service.parseRequestState(firstNonEmpty(req.StateJSON, "{}"))
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
	resp, err := h.service.Status(ctx, &pb.StatusRequest{StateJson: firstNonEmpty(req.StateJSON, "{}")})
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

func (h gopayHTTPHandler) startAuth(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.AuthStart(ctx, &pb.AuthStartRequest{Phone: firstNonEmpty(req.Phone, req.WAPhone), CountryCode: req.CountryCode, Pin: req.Pin, OtpChannel: req.OTPChannel, StateJson: firstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "login", resp.GetSuccess(), nil)
	out.applyOTP(resp.GetOtpSent(), resp.GetVerificationMethod())
	out.StateJSON = resp.GetStateJson()
	out.Phone = firstNonEmpty(req.Phone, req.WAPhone)
	out.Ready = resp.GetReady()
	out.AccountTokenReady = resp.GetReady()
	out.SignupPINComplete = !resp.GetPinSetupRequired()
	out.ErrorMessage = resp.GetErrorMessage()
	if resp.GetReady() || resp.GetOtpSent() {
		_ = h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
	}
	return out, nil
}

func (h gopayHTTPHandler) completeAuth(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.AuthComplete(ctx, &pb.AuthCompleteRequest{Otp: req.otpValue(), Pin: req.Pin, StateJson: firstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "login", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.Phone = resp.GetPhone()
	out.Ready = resp.GetReady()
	out.AccountTokenReady = resp.GetReady()
	out.SignupPINComplete = resp.GetPinSetupComplete()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfReady(ctx, req.withOTPChannel(out.OTPChannel), out)
	return out, nil
}

func (h gopayHTTPHandler) startPIN(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CreatePinStart(ctx, &pb.CreatePinStartRequest{Pin: req.Pin, OtpChannel: req.OTPChannel, StateJson: firstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "ensure_pin_setup", resp.GetSuccess(), nil)
	out.applyOTP(resp.GetOtpSent(), resp.GetVerificationMethod())
	out.StateJSON = resp.GetStateJson()
	out.ErrorMessage = resp.GetErrorMessage()
	state := h.service.parseRequestState(out.StateJSON)
	ready := resp.GetSuccess() && stateString(state, "stage") == "ready"
	out.SignupPINComplete = ready
	out.AccountTokenReady = ready
	out.Ready = ready
	_ = h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
	return out, nil
}

func (h gopayHTTPHandler) retryPIN(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CreatePinRetry(ctx, &pb.CreatePinRetryRequest{StateJson: firstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "ensure_pin_setup", resp.GetSuccess(), nil)
	out.applyOTP(resp.GetOtpSent(), "")
	out.StateJSON = resp.GetStateJson()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
	return out, nil
}

func (h gopayHTTPHandler) completePIN(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CreatePinComplete(ctx, &pb.CreatePinCompleteRequest{Otp: req.otpValue(), Pin: req.Pin, StateJson: firstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "ensure_pin_setup", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.Phone = resp.GetPhone()
	out.SignupPINComplete = resp.GetPinSetupComplete()
	out.AccountTokenReady = resp.GetPinSetupComplete()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfReady(ctx, req, out)
	return out, nil
}

func (h gopayHTTPHandler) acquireSignupPhone(req gopayActionRequest) *gopayActionResult {
	phone := firstNonEmpty(req.Phone, req.WAPhone, mapString(req.Data, "phone"))
	countryCode := firstNonEmpty(req.CountryCode, mapString(req.Data, "country_code"), "62")
	data := map[string]any{
		"phone":        phone,
		"country_code": countryCode,
	}
	out := h.baseResult(req, "gopay_app_signup_phone", phone != "", data)
	out.Phone = phone
	out.CountryCode = countryCode
	out.ActivationID = req.ActivationID
	if phone == "" {
		out.ErrorMessage = "signup phone is required"
	}
	return out
}

func (h gopayHTTPHandler) checkSignupPhone(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CheckPhone(ctx, &pb.CheckPhoneRequest{
		Phone:       req.Phone,
		CountryCode: firstNonEmpty(req.CountryCode, "62"),
		StateJson:   firstNonEmpty(req.StateJSON, mapString(req.Data, "state_json")),
	})
	if err != nil {
		return nil, err
	}
	success := resp.GetAvailable() && resp.GetStatus() != "error" && resp.GetStatus() != "rate_limited"
	data := map[string]any{
		"available":          resp.GetAvailable(),
		"status":             resp.GetStatus(),
		"proxy_hash":         resp.GetProxyHash(),
		"device_fingerprint": resp.GetDeviceFingerprint(),
		"state_json":         resp.GetStateJson(),
	}
	out := h.baseResult(req, "gopay_app_check_phone", success, data)
	out.Phone = req.Phone
	out.StateJSON = firstNonEmpty(resp.GetStateJson(), req.StateJSON, "{}")
	out.PhoneAccepted = success
	out.RotatableFailure = resp.GetStatus() == "registered"
	out.RetryableFailure = resp.GetStatus() == "error" || resp.GetStatus() == "rate_limited"
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) discardSignupPhone(req gopayActionRequest) *gopayActionResult {
	data := map[string]any{
		"activation_id": req.ActivationID,
		"reason":        strings.TrimSpace(req.Reason),
	}
	return h.baseResult(req, "gopay_app_signup_phone_cancel", true, data)
}

func (h gopayHTTPHandler) startSignup(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.SignupStart(ctx, &pb.SignupStartRequest{Phone: req.Phone, CountryCode: req.CountryCode, OtpChannel: req.OTPChannel, StateJson: firstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "signup", resp.GetSuccess(), nil)
	out.applyOTP(resp.GetOtpSent(), resp.GetVerificationMethod())
	out.StateJSON = resp.GetStateJson()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
	return out, nil
}

func (h gopayHTTPHandler) retrySignup(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.SignupRetry(ctx, &pb.SignupRetryRequest{StateJson: firstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "signup", resp.GetSuccess(), nil)
	out.applyOTP(resp.GetOtpSent(), "")
	out.StateJSON = resp.GetStateJson()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
	return out, nil
}

func (h gopayHTTPHandler) completeSignup(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.SignupComplete(ctx, &pb.SignupCompleteRequest{Otp: req.otpValue(), StateJson: firstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "signup", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.Phone = resp.GetPhone()
	out.SignupComplete = resp.GetSuccess()
	out.RequirePIN = resp.GetPinSetupRequired()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfReady(ctx, req.withOTPChannel(out.OTPChannel), out)
	return out, nil
}

func (h gopayHTTPHandler) startDeactivate(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.DeactivateStart(ctx, &pb.DeactivateStartRequest{Pin: req.Pin, StateJson: firstNonEmpty(req.StateJSON, "{}")})
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
	resp, err := h.service.DeactivateComplete(ctx, &pb.DeactivateCompleteRequest{Otp: req.otpValue(), StateJson: firstNonEmpty(req.StateJSON, "{}")})
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

func (h gopayHTTPHandler) acquireChangePhone(req gopayActionRequest) *gopayActionResult {
	out := h.baseResult(req, "change_phone_get_number", strings.TrimSpace(req.Phone) != "", nil)
	out.Phone = req.Phone
	out.ActivationID = req.ActivationID
	if out.Phone == "" {
		out.ErrorMessage = "change phone target phone is required"
	}
	return out
}

func (h gopayHTTPHandler) startChangePhone(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.ChangePhoneStart(ctx, &pb.ChangePhoneStartRequest{Pin: req.Pin, NewPhone: req.Phone, CountryCode: req.CountryCode, StateJson: firstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "change_phone_start", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.Phone = resp.GetNewPhone()
	out.ActivationID = resp.GetActivationId()
	out.applyOTP(resp.GetOtpSent(), "sms")
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) retryChangePhone(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.ChangePhoneRetry(ctx, &pb.ChangePhoneRetryRequest{StateJson: firstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "change_phone_retry", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.applyOTP(resp.GetOtpSent(), "sms")
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) completeChangePhone(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.ChangePhoneComplete(ctx, &pb.ChangePhoneCompleteRequest{Otp: req.otpValue(), StateJson: firstNonEmpty(req.StateJSON, "{}")})
	if err != nil {
		return nil, err
	}
	out := h.baseResult(req, "change_phone_complete", resp.GetSuccess(), nil)
	out.StateJSON = resp.GetStateJson()
	out.ChangePhoneComplete = resp.GetSuccess()
	out.ErrorMessage = resp.GetErrorMessage()
	_ = h.saveStateIfAccount(ctx, req, out.StateJSON)
	return out, nil
}

func (h gopayHTTPHandler) awaitChannelOTP(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	otp := req.otpValue()
	channel := normalizeActionOTPChannel(firstNonEmpty(req.Channel, req.OTPChannel))
	if channel == "" {
		channel = "wa"
	}
	if otp == "" {
		latest, found, err := h.latestChannelOTP(ctx, req, channel)
		if err != nil {
			return nil, err
		}
		if found {
			otp = latest.OTP
			req.OTPSource = firstNonEmpty(req.OTPSource, latest.Source)
		}
	}
	if otp == "" {
		if err := h.registerChannelOTPWait(ctx, firstNonEmpty(req.actionScope, gopayAccountActionScope), "await_channel_otp", req, channel); err != nil {
			return nil, err
		}
	}
	data := map[string]any{"otp_channel": channel, "channel_otp_target": strings.TrimSpace(req.Target), "otp_found": otp != "", "otp_issued_after_unix": req.OTPIssuedAfter}
	if otp != "" {
		data["otp"] = otp
	}
	out := h.baseResult(req, "await_channel_otp", true, data)
	out.OTPChannel = channel
	out.OTPFound = otp != ""
	if otp != "" {
		out.OTPSource = firstNonEmpty(req.OTPSource, channel)
	}
	out.OTPIssuedAfterUnix = req.OTPIssuedAfter
	out.OTPTimeoutSeconds = req.OTPTimeoutSeconds
	return out, nil
}

func (h gopayHTTPHandler) awaitPaymentChannelOTP(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	otp := req.otpValue()
	channel := normalizeActionOTPChannel(firstNonEmpty(req.Channel, req.OTPChannel))
	if channel == "" {
		channel = "wa"
	}
	if otp == "" {
		latest, found, err := h.latestChannelOTP(ctx, req, channel)
		if err != nil {
			return nil, err
		}
		if found {
			otp = latest.OTP
			req.OTPSource = firstNonEmpty(req.OTPSource, latest.Source)
		}
	}
	if otp == "" {
		if err := h.registerChannelOTPWait(ctx, scope, "await_channel_otp", req, channel); err != nil {
			return nil, err
		}
	}
	data := map[string]any{
		"otp_channel":           channel,
		"channel_otp_target":    strings.TrimSpace(req.Target),
		"otp_found":             otp != "",
		"otp_issued_after_unix": req.OTPIssuedAfter,
		"runtime_owner":         "gopay-app",
	}
	if otp != "" {
		data["otp"] = otp
	}
	out := h.basePaymentResult(scope, req, "await_channel_otp", true, data)
	out.OTPChannel = channel
	out.OTPFound = otp != ""
	if otp != "" {
		out.OTPSource = firstNonEmpty(req.OTPSource, channel)
	}
	out.OTPIssuedAfterUnix = req.OTPIssuedAfter
	out.OTPTimeoutSeconds = req.OTPTimeoutSeconds
	return out, nil
}

func (h gopayHTTPHandler) latestChannelOTP(ctx context.Context, req gopayActionRequest, channel string) (latestChannelOTP, bool, error) {
	target := firstNonEmpty(req.Target, req.ActivationID, req.GopayAccountID)
	timeout := firstNonZeroInt32(req.OTPTimeoutSeconds, defaultChannelOTPTimeoutSeconds(channel))
	return h.service.store.LatestChannelOTP(ctx, channel, target, req.OTPIssuedAfter, timeout)
}

func (h gopayHTTPHandler) registerChannelOTPWait(ctx context.Context, scope string, step string, req gopayActionRequest, channel string) error {
	target := firstNonEmpty(req.Target, req.ActivationID, req.GopayAccountID)
	entry := channelOTPWaitEntry{
		JobID:           req.JobID,
		AccountID:       req.GopayAccountID,
		N8NExecutionID:  req.N8NExecutionID,
		Action:          scope,
		StepName:        step,
		Channel:         channel,
		Target:          target,
		IssuedAfterUnix: req.OTPIssuedAfter,
		TimeoutSeconds:  firstNonZeroInt32(req.OTPTimeoutSeconds, defaultChannelOTPTimeoutSeconds(channel)),
		ResumeURL:       req.ResumeURL,
	}
	return h.service.store.RegisterChannelOTPWait(ctx, entry, channelOTPWaitTTL(entry.TimeoutSeconds))
}

func (h gopayHTTPHandler) startPayment(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	tokenization := firstNonEmpty(req.Tokenization, "true")
	resp, err := h.service.StartGopayPayment(ctx, &pb.StartGopayPaymentRequest{
		SnapToken:         req.SnapToken,
		CheckoutUrl:       req.CheckoutURL,
		CheckoutSessionId: req.CheckoutSessionID,
		Tokenization:      tokenization,
		GopayAccountId:    req.GopayAccountID,
		Amount:            req.Amount,
		Currency:          req.Currency,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		resp = &pb.StartGopayPaymentResponse{Success: false, ErrorMessage: "empty gopay payment start response"}
	}
	data := paymentResponseData(resp, nil)
	data["runtime_owner"] = "gopay-app"
	out := h.basePaymentResult(scope, req, "gopay_payment", resp.GetSuccess(), data)
	out.GopayAccountID = resp.GetGopayAccountId()
	out.OTPChannel = resp.GetOtpChannel()
	out.FlowID = resp.GetFlowId()
	out.CheckoutURL = resp.GetCheckoutUrl()
	out.CheckoutSessionID = resp.GetCheckoutSessionId()
	out.UseAccountToken = req.UseAccountToken
	out.OTPRequired = resp.GetOtpRequired()
	out.OTPIssuedAfterUnix = resp.GetIssuedAfterUnix()
	out.OTPTimeoutSeconds = firstNonZeroInt32(req.OTPTimeoutSeconds, 300)
	out.SnapTokenPresent = resp.GetSnapToken() != ""
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) completePayment(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CompleteGopayPayment(ctx, &pb.CompleteGopayPaymentRequest{
		FlowId: req.FlowID,
		Otp:    req.otpValue(),
	})
	if err != nil {
		return nil, err
	}
	return h.paymentResult(scope, req, "gopay_payment", resp), nil
}

func (h gopayHTTPHandler) confirmPayment(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.ConfirmGopayPayment(ctx, &pb.ConfirmGopayPaymentRequest{FlowId: req.FlowID})
	if err != nil {
		return nil, err
	}
	return h.paymentResult(scope, req, "gopay_payment", resp), nil
}

func (h gopayHTTPHandler) cancelPayment(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.CancelGopayPayment(ctx, &pb.CancelGopayPaymentRequest{FlowId: req.FlowID})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		resp = &pb.CancelGopayPaymentResponse{Success: false, ErrorMessage: "empty gopay payment cancel response"}
	}
	data := map[string]any{"flow_id": strings.TrimSpace(req.FlowID), "runtime_owner": "gopay-app"}
	out := h.basePaymentResult(scope, req, "cancel_payment", resp.GetSuccess(), data)
	out.FlowID = strings.TrimSpace(req.FlowID)
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) resendPaymentOTP(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	resp, err := h.service.ResendGopayPaymentOTP(ctx, &pb.ResendGopayPaymentOTPRequest{FlowId: req.FlowID})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		resp = &pb.ResendGopayPaymentOTPResponse{Success: false, ErrorMessage: "empty gopay payment otp response"}
	}
	data := map[string]any{
		"response_present":      true,
		"success":               resp.GetSuccess(),
		"error_message":         resp.GetErrorMessage(),
		"flow_id":               resp.GetFlowId(),
		"otp_issued_after_unix": resp.GetIssuedAfterUnix(),
		"runtime_owner":         "gopay-app",
	}
	out := h.basePaymentResult(scope, req, "resend_payment_otp", resp.GetSuccess(), data)
	out.FlowID = resp.GetFlowId()
	out.OTPIssuedAfterUnix = resp.GetIssuedAfterUnix()
	out.OTPRequired = true
	out.OTPTimeoutSeconds = firstNonZeroInt32(req.OTPTimeoutSeconds, 300)
	out.ErrorMessage = resp.GetErrorMessage()
	return out, nil
}

func (h gopayHTTPHandler) paymentResult(scope string, req gopayActionRequest, step string, resp *pb.GopayPaymentResponse) *gopayActionResult {
	if resp == nil {
		resp = &pb.GopayPaymentResponse{Success: false, ErrorMessage: "empty gopay payment response"}
	}
	data := paymentResponseData(nil, resp)
	data["runtime_owner"] = "gopay-app"
	out := h.basePaymentResult(scope, req, step, resp.GetSuccess(), data)
	out.FlowID = req.FlowID
	out.ChargeRef = resp.GetChargeRef()
	out.SnapTokenPresent = resp.GetSnapToken() != ""
	out.AwaitingManual = resp.GetAwaitingManualConfirmation()
	settled := resp.GetSuccess() && !resp.GetAwaitingManualConfirmation() && resp.GetChargeRef() != ""
	out.PlusTrialEligible = settled
	out.PlusTrialChecked = settled
	out.PlusActive = settled
	out.ErrorMessage = resp.GetErrorMessage()
	return out
}

func (h gopayHTTPHandler) finish(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	if err := h.saveStateIfAccount(ctx, req, req.StateJSON); err != nil {
		return nil, err
	}
	job, _ := h.loadWorkflowJob(ctx, req.JobID)
	if job != nil {
		job.Status = "succeeded"
		job.Result = req.Data
		job.UpdatedAtUnix = time.Now().Unix()
		_ = h.saveWorkflowJob(ctx, job)
	}
	return h.baseResult(req, "finish", true, req.Data), nil
}

func (h gopayHTTPHandler) fail(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	job, _ := h.loadWorkflowJob(ctx, req.JobID)
	if job != nil {
		job.Status = "failed"
		job.ErrorMessage = req.ErrorMessage
		job.Result = req.Data
		job.UpdatedAtUnix = time.Now().Unix()
		_ = h.saveWorkflowJob(ctx, job)
	}
	out := h.baseResult(req, "fail", false, req.Data)
	out.ErrorMessage = req.ErrorMessage
	return out, nil
}

func (h gopayHTTPHandler) genericSuccess(req gopayActionRequest, step string, data map[string]any) *gopayActionResult {
	return h.baseResult(req, step, true, data)
}

func (h gopayHTTPHandler) baseResult(req gopayActionRequest, step string, success bool, data map[string]any) *gopayActionResult {
	return &gopayActionResult{JobID: req.JobID, N8NExecutionID: req.N8NExecutionID, Action: firstNonEmpty(req.actionScope, gopayAccountActionScope), Step: step, Operation: normalizeGoPayWorkflowOperation(req.Operation), Success: success, GopayAccountID: req.GopayAccountID, Phone: req.Phone, OTPChannel: normalizeActionOTPChannel(req.OTPChannel), CountryCode: req.CountryCode, ActivationID: req.ActivationID, FailureCount: req.FailureCount, StateJSON: firstNonEmpty(req.StateJSON, "{}"), Data: data}
}

func (h gopayHTTPHandler) basePaymentResult(scope string, req gopayActionRequest, step string, success bool, data map[string]any) *gopayActionResult {
	return &gopayActionResult{JobID: req.JobID, N8NExecutionID: req.N8NExecutionID, Action: strings.TrimSpace(scope), Step: step, Success: success, GopayAccountID: req.GopayAccountID, Phone: firstNonEmpty(req.Phone, req.WAPhone), OTPChannel: normalizeActionOTPChannel(firstNonEmpty(req.OTPChannel, req.Channel)), CountryCode: req.CountryCode, ActivationID: req.ActivationID, StateJSON: firstNonEmpty(req.StateJSON, "{}"), Data: data}
}

func (r *gopayActionRequest) withJob(job *gopayWorkflowJob) gopayActionRequest {
	if job == nil {
		r.Operation = normalizeGoPayWorkflowOperation(r.Operation)
		r.OTPChannel = normalizeActionOTPChannel(r.OTPChannel)
		return *r
	}
	r.JobID = firstNonEmpty(r.JobID, job.JobID)
	r.N8NExecutionID = firstNonEmpty(r.N8NExecutionID, job.N8NExecutionID)
	r.Operation = normalizeGoPayWorkflowOperation(firstNonEmpty(r.Operation, job.Operation))
	r.GopayAccountID = goPayAppAccountID(firstNonEmpty(r.GopayAccountID, job.GopayAccountID))
	r.Phone = firstNonEmpty(r.Phone, job.Phone)
	r.CountryCode = firstNonEmpty(r.CountryCode, job.CountryCode)
	r.Pin = firstNonEmpty(r.Pin, job.Pin)
	r.OTPChannel = normalizeActionOTPChannel(firstNonEmpty(r.OTPChannel, job.OTPChannel))
	return *r
}

func (r gopayActionRequest) withOTPChannel(channel string) gopayActionRequest {
	if normalized := normalizeActionOTPChannel(channel); normalized != "" {
		r.OTPChannel = normalized
	}
	return r
}

func (r gopayActionRequest) otpValue() string {
	for _, value := range []string{r.OTP, mapString(r.Data, "otp"), mapString(r.Data, "channel_otp"), mapString(r.Data, "code")} {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func (r *gopayActionResult) applyOTP(sent bool, method string) {
	r.OTPSent = sent
	r.OTPRequired = sent
	r.OTPIssuedAfterUnix = time.Now().Unix()
	r.OTPTimeoutSeconds = 300
	if channel := normalizeActionOTPChannel(method); channel != "" {
		r.OTPChannel = channel
	}
}

func (h gopayHTTPHandler) saveStateIfReady(ctx context.Context, req gopayActionRequest, out *gopayActionResult) error {
	if out == nil || (!out.Ready && !out.AccountTokenReady && !out.SignupPINComplete && !out.SignupComplete) {
		return nil
	}
	return h.saveStateIfAccount(ctx, req.withOTPChannel(out.OTPChannel), out.StateJSON)
}

func (h gopayHTTPHandler) saveStateIfAccount(ctx context.Context, req gopayActionRequest, rawStateJSON string) error {
	if strings.TrimSpace(req.GopayAccountID) == "" || strings.TrimSpace(rawStateJSON) == "" || strings.TrimSpace(rawStateJSON) == "{}" {
		return nil
	}
	state := h.service.parseRequestState(rawStateJSON)
	if channel := normalizeActionOTPChannel(req.OTPChannel); channel != "" {
		state["_otp_channel"] = channel
	}
	persistGopayAccountOTPChannel(state)
	rawStateJSON = stateJSON(state)
	_, err := h.service.SaveGopayAccountState(ctx, &pb.SaveGopayAccountStateRequest{GopayAccountId: req.GopayAccountID, StateJson: rawStateJSON})
	if err != nil {
		return err
	}
	return h.syncAccountProfile(ctx, req, state)
}

func (h gopayHTTPHandler) syncAccountProfile(ctx context.Context, req gopayActionRequest, state stateMap) error {
	if strings.TrimSpace(req.GopayAccountID) == "" {
		return nil
	}
	phone := firstNonEmpty(
		req.WAPhone,
		req.Phone,
		stateString(state, "phone"),
		stateString(state, "_login_phone"),
		stateString(state, "_signup_phone"),
	)
	countryCode := firstNonEmpty(req.CountryCode, stateCountryCode(state))
	otpChannel := normalizeActionOTPChannel(firstNonEmpty(
		req.OTPChannel,
		gopayAccountOTPChannelFromState(state),
	))
	if strings.TrimSpace(phone) == "" && strings.TrimSpace(countryCode) == "" && strings.TrimSpace(req.Pin) == "" && otpChannel == "" {
		return nil
	}
	_, err := h.service.SaveGopayAccountProfile(ctx, &pb.SaveGopayAccountProfileRequest{
		GopayAccountId: req.GopayAccountID,
		WaPhone:        phone,
		CountryCode:    countryCode,
		OtpChannel:     otpChannel,
		Pin:            req.Pin,
	})
	return err
}

func readGopayActionRequest(w http.ResponseWriter, r *http.Request) (gopayActionRequest, error) {
	var req gopayActionRequest
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		return req, err
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return req, nil
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return req, err
	}
	return req, nil
}

func newGoPayWorkflowJob(req *pb.GoPayAccountWorkflowRequest) (*gopayWorkflowJob, error) {
	operation := normalizeGoPayWorkflowOperation(req.GetOperation().String())
	accountID := goPayAppAccountID(req.GetGopayAccountId())
	if accountID == "" {
		return nil, fmt.Errorf("gopay_account_id is required")
	}
	now := time.Now().Unix()
	return &gopayWorkflowJob{JobID: uuid.NewString(), Operation: operation, GopayAccountID: accountID, Phone: strings.TrimSpace(req.GetPhone()), CountryCode: strings.TrimSpace(req.GetCountryCode()), Pin: strings.TrimSpace(req.GetPin()), OTPChannel: normalizeActionOTPChannel(req.GetOtpChannel()), Status: "started", CreatedAtUnix: now, UpdatedAtUnix: now}, nil
}

func newGoPayRegisterIndonesiaWAWorkflowJob() *gopayWorkflowJob {
	now := time.Now().Unix()
	return &gopayWorkflowJob{JobID: uuid.NewString(), Operation: registerIndonesiaWAWorkflowOperation, Status: "started", CreatedAtUnix: now, UpdatedAtUnix: now}
}

func (h gopayHTTPHandler) triggerGoPayWorkflow(ctx context.Context, jobID string, webhookPath string, workflowName string) error {
	if h.n8nWebhookBaseURL == "" {
		return fmt.Errorf("GOPAY_N8N_WEBHOOK_BASE_URL is required")
	}
	url := h.n8nWebhookBaseURL + "/" + strings.Trim(strings.TrimSpace(webhookPath), "/")
	body, _ := json.Marshal(map[string]string{"job_id": strings.TrimSpace(jobID)})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("n8n %s workflow returned HTTP %d", strings.TrimSpace(workflowName), resp.StatusCode)
	}
	return nil
}

func (h gopayHTTPHandler) triggerGoPayWorkflowAsync(job *gopayWorkflowJob, webhookPath string, workflowName string) {
	if job == nil {
		return
	}
	go func(snapshot gopayWorkflowJob) {
		if err := h.triggerGoPayWorkflow(context.Background(), snapshot.JobID, webhookPath, workflowName); err != nil {
			snapshot.Status = "trigger_failed"
			snapshot.ErrorMessage = err.Error()
			snapshot.UpdatedAtUnix = time.Now().Unix()
			_ = h.saveWorkflowJob(context.Background(), &snapshot)
		}
	}(*job)
}

func (h gopayHTTPHandler) loadWorkflowJob(ctx context.Context, jobID string) (*gopayWorkflowJob, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	raw, err := h.service.store.Load(ctx, workflowJobKey(jobID))
	if err != nil {
		return nil, err
	}
	var job gopayWorkflowJob
	if err := json.Unmarshal([]byte(firstNonEmpty(raw, "{}")), &job); err != nil {
		return nil, err
	}
	if strings.TrimSpace(job.JobID) == "" {
		return nil, fmt.Errorf("gopay workflow job not found: %s", jobID)
	}
	return &job, nil
}

func (h gopayHTTPHandler) saveWorkflowJob(ctx context.Context, job *gopayWorkflowJob) error {
	if job == nil || strings.TrimSpace(job.JobID) == "" {
		return fmt.Errorf("job_id is required")
	}
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = h.service.store.Save(ctx, workflowJobKey(job.JobID), string(data))
	return err
}

func workflowJobKey(jobID string) string {
	return "workflow-job:" + strings.TrimSpace(jobID)
}

func goPayAppAccountID(value string) string {
	return strings.TrimSpace(value)
}

func normalizeActionOTPChannel(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "wa", "whatsapp", "otp_wa", "gopay_otp_channel_whatsapp":
		return "wa"
	case "sms", "otp_sms", "gopay_otp_channel_sms":
		return "sms"
	default:
		return value
	}
}

func normalizeGoPayWorkflowOperation(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "gopay_account_workflow_operation_")
	switch value {
	case "", "unspecified":
		return "login"
	case "ensure_pin_setup":
		return "ensure_pin_setup"
	case "check_balance":
		return "check_balance"
	case "check_pin":
		return "check_pin"
	case "change_phone":
		return "change_phone"
	case "provision":
		return "provision"
	case "deactivate":
		return "deactivate"
	case "signup":
		return "signup"
	default:
		return value
	}
}

func mapString(data map[string]any, key string) string {
	if data == nil {
		return ""
	}
	if value, ok := data[key]; ok {
		return strings.TrimSpace(fmt.Sprint(value))
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonZeroInt32(values ...int32) int32 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func paymentResponseData(start *pb.StartGopayPaymentResponse, result *pb.GopayPaymentResponse) map[string]any {
	data := map[string]any{"response_present": start != nil || result != nil}
	if start != nil {
		data["success"] = start.GetSuccess()
		data["error_message"] = start.GetErrorMessage()
		data["flow_id"] = start.GetFlowId()
		data["snap_token_present"] = start.GetSnapToken() != ""
		data["issued_after_unix"] = start.GetIssuedAfterUnix()
		data["expires_at_unix"] = start.GetExpiresAtUnix()
		data["checkout_url"] = start.GetCheckoutUrl()
		data["checkout_session_id"] = start.GetCheckoutSessionId()
		data["otp_required"] = start.GetOtpRequired()
		data["otp_channel"] = start.GetOtpChannel()
		data["gopay_account_id"] = start.GetGopayAccountId()
		data["amount"] = start.GetAmount()
		data["currency"] = start.GetCurrency()
		data["otp_target"] = start.GetOtpTarget()
	}
	if result != nil {
		data["success"] = result.GetSuccess()
		data["error_message"] = result.GetErrorMessage()
		data["charge_ref"] = result.GetChargeRef()
		data["snap_token_present"] = result.GetSnapToken() != ""
		data["awaiting_manual_confirmation"] = result.GetAwaitingManualConfirmation()
		data["deeplink_url"] = result.GetDeeplinkUrl()
		data["qr_code_url"] = result.GetQrCodeUrl()
		data["qr_string"] = result.GetQrString()
		data["finish_redirect_url"] = result.GetFinishRedirectUrl()
		data["finish_200_redirect_url"] = result.GetFinish_200RedirectUrl()
	}
	return data
}
