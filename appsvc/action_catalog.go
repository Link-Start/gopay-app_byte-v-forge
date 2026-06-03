package appsvc

import (
	"context"

	"github.com/byte-v-forge/common-lib/accountaction"
	accountv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/account/v1"
	"github.com/byte-v-forge/gopay-app/pb"
)

const (
	ActionGoPaySignup       = "GOPAY_ACCOUNT_SIGNUP"
	ActionGoPayLogin        = "GOPAY_ACCOUNT_LOGIN"
	ActionGoPayEnsurePIN    = "GOPAY_ACCOUNT_ENSURE_PIN_SETUP"
	ActionGoPayCheckBalance = "GOPAY_ACCOUNT_CHECK_BALANCE"
	ActionGoPayCheckPIN     = "GOPAY_ACCOUNT_CHECK_PIN"
	ActionGoPayChangePhone  = "GOPAY_ACCOUNT_CHANGE_PHONE"
	ActionGoPayDeactivate   = "GOPAY_ACCOUNT_DEACTIVATE"

	CapabilityGoPay         = "gopay"
	CapabilityGoPayAccount  = "gopay_account"
	CapabilityN8NWorkflow   = "n8n_workflow"
	AccountActionVisibility = "gopay_account"
	AccountActionPlacement  = "gopay"
)

func (s *Server) GetActionCatalog(context.Context, *pb.GetGopayActionCatalogRequest) (*pb.GetGopayActionCatalogResponse, error) {
	return &pb.GetGopayActionCatalogResponse{Success: true, Catalog: GoPayActionCatalog()}, nil
}

func GoPayActionCatalog() *accountv1.AccountActionCatalog {
	return accountaction.Catalog(
		goPayWorkflowAction(ActionGoPaySignup, "GoPay 注册", goPayAccountSignupWorkflow,
			accountaction.RequiredFields("account_id", "phone", "country_code", "otp_channel"),
		),
		goPayWorkflowAction(ActionGoPayLogin, "GoPay 登录", goPayAccountLoginWorkflow,
			accountaction.RequiredFields("account_id", "phone", "country_code", "otp_channel"),
		),
		goPayWorkflowAction(ActionGoPayEnsurePIN, "GoPay PIN 设置", goPayAccountEnsurePINWorkflow,
			accountaction.RequiredFields("account_id", "phone", "country_code", "otp_channel"),
		),
		goPayWorkflowAction(ActionGoPayCheckBalance, "GoPay 查余额", goPayAccountCheckBalanceWorkflow,
			accountaction.RequiredFields("account_id", "phone", "country_code"),
		),
		goPayWorkflowAction(ActionGoPayCheckPIN, "GoPay 查 PIN", goPayAccountCheckPINWorkflow,
			accountaction.RequiredFields("account_id"),
		),
		goPayWorkflowAction(ActionGoPayChangePhone, "GoPay 改绑手机号", goPayAccountChangePhoneWorkflow,
			accountaction.RequiredFields("account_id", "phone", "country_code"),
		),
		goPayWorkflowAction(ActionGoPayDeactivate, "GoPay 注销", goPayAccountDeactivateWorkflow,
			accountaction.RequiredFields("account_id", "phone", "country_code"),
		),
	)
}

func goPayWorkflowAction(actionID, displayName string, workflow goPayAccountWorkflow, options ...accountaction.DefinitionOption) *accountv1.AccountActionDefinition {
	defOptions := []accountaction.DefinitionOption{
		accountaction.Owner("gopay-app"),
		accountaction.Visibility(AccountActionVisibility),
		accountaction.RequestProto("gopay_app.GoPayAccountWorkflowRequest"),
		accountaction.ResponseProto("gopay_app.GoPayAccountWorkflowResponse"),
		accountaction.N8NWorkflow(
			workflow.Key,
			workflow.Key+"-",
			"/workflows/"+workflow.Key,
			"gopay-account",
			workflow.WebhookPath,
			"/actions/gopay-account",
			accountv1.AccountActionAPIKind_ACCOUNT_ACTION_API_KIND_RAW_N8N,
		),
		accountaction.DefaultButton(workflow.ButtonLabel, AccountActionPlacement, accountaction.ButtonIntent(workflow.Intent)),
		accountaction.Capabilities(CapabilityGoPay, CapabilityGoPayAccount, CapabilityN8NWorkflow),
	}
	defOptions = append(defOptions, options...)
	return accountaction.Definition(actionID, displayName, defOptions...)
}
