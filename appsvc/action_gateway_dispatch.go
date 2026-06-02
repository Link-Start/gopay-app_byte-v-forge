package appsvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) invokeGoPayAccountAction(ctx context.Context, action string, req gopayActionRequest) (*gopayActionResult, error) {
	job, loadErr := h.loadWorkflowJob(ctx, req.JobID)
	if strings.TrimSpace(req.JobID) != "" && job == nil && strings.TrimSpace(action) != "fail" {
		return nil, loadErr
	}
	req = req.withJob(job)
	return h.invokeGoPayAccountRuntimeAction(ctx, action, req, job)
}

func (h gopayHTTPHandler) invokeGoPayAccountRuntimeAction(ctx context.Context, action string, req gopayActionRequest, job *pb.GopayWorkflowJob) (*gopayActionResult, error) {
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
