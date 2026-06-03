package appsvc

import "strings"

type goPayAccountWorkflow struct {
	Key         string
	Operation   string
	WebhookPath string
	Label       string
	ButtonLabel string
	Intent      string
}

var (
	goPayAccountLoginWorkflow = goPayAccountWorkflow{
		Key:         "gopay-account-login",
		Operation:   "login",
		WebhookPath: "gopay-app/account/login",
		Label:       "GoPay 登录",
		ButtonLabel: "登录",
		Intent:      "login",
	}
	goPayAccountSignupWorkflow = goPayAccountWorkflow{
		Key:         "gopay-account-signup",
		Operation:   "signup",
		WebhookPath: "gopay-app/account/signup",
		Label:       "GoPay 注册",
		ButtonLabel: "注册",
		Intent:      "signup",
	}
	goPayAccountEnsurePINWorkflow = goPayAccountWorkflow{
		Key:         "gopay-account-ensure-pin",
		Operation:   "ensure_pin_setup",
		WebhookPath: "gopay-app/account/ensure-pin",
		Label:       "GoPay PIN 设置",
		ButtonLabel: "PIN 设置",
		Intent:      "ensure_pin_setup",
	}
	goPayAccountCheckBalanceWorkflow = goPayAccountWorkflow{
		Key:         "gopay-account-check-balance",
		Operation:   "check_balance",
		WebhookPath: "gopay-app/account/check-balance",
		Label:       "GoPay 查余额",
		ButtonLabel: "查余额",
		Intent:      "check_balance",
	}
	goPayAccountCheckPINWorkflow = goPayAccountWorkflow{
		Key:         "gopay-account-check-pin",
		Operation:   "check_pin",
		WebhookPath: "gopay-app/account/check-pin",
		Label:       "GoPay 查 PIN",
		ButtonLabel: "查 PIN",
		Intent:      "check_pin",
	}
	goPayAccountChangePhoneWorkflow = goPayAccountWorkflow{
		Key:         "gopay-account-change-phone",
		Operation:   "change_phone",
		WebhookPath: "gopay-app/account/change-phone",
		Label:       "GoPay 改绑手机号",
		ButtonLabel: "改绑手机号",
		Intent:      "change_phone",
	}
	goPayAccountDeactivateWorkflow = goPayAccountWorkflow{
		Key:         "gopay-account-deactivate",
		Operation:   "deactivate",
		WebhookPath: "gopay-app/account/deactivate",
		Label:       "GoPay 注销",
		ButtonLabel: "注销",
		Intent:      "deactivate",
	}
)

func goPayAccountWorkflows() []goPayAccountWorkflow {
	return []goPayAccountWorkflow{
		goPayAccountLoginWorkflow,
		goPayAccountSignupWorkflow,
		goPayAccountEnsurePINWorkflow,
		goPayAccountCheckBalanceWorkflow,
		goPayAccountCheckPINWorkflow,
		goPayAccountChangePhoneWorkflow,
		goPayAccountDeactivateWorkflow,
	}
}

func goPayAccountWorkflowByKey(key string) (goPayAccountWorkflow, bool) {
	key = strings.Trim(strings.TrimSpace(key), "/")
	for _, workflow := range goPayAccountWorkflows() {
		if workflow.Key == key {
			return workflow, true
		}
	}
	return goPayAccountWorkflow{}, false
}
