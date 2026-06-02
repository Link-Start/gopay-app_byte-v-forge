package appsvc

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) awaitChannelOTP(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	otp := req.otpValue()
	channel := normalizeActionOTPChannel(stringx.FirstNonEmpty(req.Channel, req.OTPChannel))
	if channel == "" {
		channel = "wa"
	}
	if otp == "" {
		latest, found, err := h.latestChannelOTP(ctx, req, channel)
		if err != nil {
			return nil, err
		}
		if found {
			otp = latest.GetOtp()
			req.OTPSource = stringx.FirstNonEmpty(req.OTPSource, latest.GetSource())
		}
	}
	if otp == "" {
		if err := h.registerChannelOTPWait(ctx, stringx.FirstNonEmpty(req.actionScope, gopayAccountActionScope), "await_channel_otp", req, channel); err != nil {
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
		out.OTPSource = stringx.FirstNonEmpty(req.OTPSource, channel)
	}
	out.OTPIssuedAfterUnix = req.OTPIssuedAfter
	out.OTPTimeoutSeconds = req.OTPTimeoutSeconds
	return out, nil
}

func (h gopayHTTPHandler) awaitPaymentChannelOTP(ctx context.Context, scope string, req gopayActionRequest) (*gopayActionResult, error) {
	otp := req.otpValue()
	channel := normalizeActionOTPChannel(stringx.FirstNonEmpty(req.Channel, req.OTPChannel))
	if channel == "" {
		channel = "wa"
	}
	if otp == "" {
		latest, found, err := h.latestChannelOTP(ctx, req, channel)
		if err != nil {
			return nil, err
		}
		if found {
			otp = latest.GetOtp()
			req.OTPSource = stringx.FirstNonEmpty(req.OTPSource, latest.GetSource())
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
		out.OTPSource = stringx.FirstNonEmpty(req.OTPSource, channel)
	}
	out.OTPIssuedAfterUnix = req.OTPIssuedAfter
	out.OTPTimeoutSeconds = req.OTPTimeoutSeconds
	return out, nil
}

func (h gopayHTTPHandler) latestChannelOTP(ctx context.Context, req gopayActionRequest, channel string) (*pb.LatestChannelOTP, bool, error) {
	target := stringx.FirstNonEmpty(req.Target, req.ActivationID, req.GopayAccountID)
	timeout := firstNonZeroInt32(req.OTPTimeoutSeconds, defaultChannelOTPTimeoutSeconds(channel))
	return h.service.store.LatestChannelOTP(ctx, channel, target, req.OTPIssuedAfter, timeout)
}

func (h gopayHTTPHandler) registerChannelOTPWait(ctx context.Context, scope string, step string, req gopayActionRequest, channel string) error {
	target := stringx.FirstNonEmpty(req.Target, req.ActivationID, req.GopayAccountID)
	entry := &pb.ChannelOTPWaitEntry{
		JobId:           req.JobID,
		AccountId:       req.GopayAccountID,
		N8NExecutionId:  req.N8NExecutionID,
		Action:          scope,
		StepName:        step,
		Channel:         channel,
		Target:          target,
		IssuedAfterUnix: req.OTPIssuedAfter,
		TimeoutSeconds:  firstNonZeroInt32(req.OTPTimeoutSeconds, defaultChannelOTPTimeoutSeconds(channel)),
		ResumeUrl:       req.ResumeURL,
	}
	return h.service.store.RegisterChannelOTPWait(ctx, entry, channelOTPWaitTTL(entry.GetTimeoutSeconds()))
}
