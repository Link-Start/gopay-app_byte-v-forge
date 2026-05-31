package appsvc

import (
	"context"

	"github.com/byte-v-forge/common-lib/accountaction"
	accountv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/account/v1"
	"github.com/byte-v-forge/gopay-app/pb"
)

const (
	ActionGoPayAccount = "GOPAY_ACCOUNT"

	CapabilityGoPay        = "gopay"
	CapabilityGoPayAccount = "gopay_account"
	CapabilityN8NWorkflow  = "n8n_workflow"

	AccountActionVisibility = "gopay_account"
	AccountActionPlacement  = "gopay"
)

func (s *Server) GetActionCatalog(context.Context, *pb.GetGopayActionCatalogRequest) (*pb.GetGopayActionCatalogResponse, error) {
	return &pb.GetGopayActionCatalogResponse{Success: true, Catalog: GoPayActionCatalog()}, nil
}

func GoPayActionCatalog() *accountv1.AccountActionCatalog {
	return accountaction.Catalog(goPayAccountWorkflowAction())
}

func goPayAccountWorkflowAction() *accountv1.AccountActionDefinition {
	return accountaction.Definition(
		ActionGoPayAccount,
		"GoPay Account",
		accountaction.Owner("gopay-app"),
		accountaction.Visibility(AccountActionVisibility),
		accountaction.RequestProto("gopay_app.GoPayAccountWorkflowRequest"),
		accountaction.ResponseProto("gopay_app.GoPayAccountWorkflowResponse"),
		accountaction.N8NWorkflow(
			"gopay-account",
			"gopay-account-",
			"/workflows/gopay-account",
			"gopay-account",
			"gopay-app/account",
			"/actions/gopay-account",
			accountv1.AccountActionAPIKind_ACCOUNT_ACTION_API_KIND_RAW_N8N,
		),
		accountaction.DefaultButton("GoPay Account", AccountActionPlacement),
		accountaction.Capabilities(CapabilityGoPay, CapabilityGoPayAccount, CapabilityN8NWorkflow),
	)
}
