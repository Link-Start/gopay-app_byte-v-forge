package appsvc

import "context"

const (
	gopayAccountActionScope                 = "gopay-account"
	gopayToolboxActionScope                 = "gopay-toolbox"
	gopayPaymentActionScope                 = "gopay-payment"
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

type gopayActionInvoker func(context.Context, string, gopayActionRequest) (*gopayActionResult, error)
