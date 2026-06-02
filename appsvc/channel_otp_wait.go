package appsvc

import (
	"strings"
	"time"
	"unicode"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func normalizeChannelOTPWaitEntry(entry *pb.ChannelOTPWaitEntry) *pb.ChannelOTPWaitEntry {
	if entry == nil {
		entry = &pb.ChannelOTPWaitEntry{}
	}
	entry.JobId = strings.TrimSpace(entry.GetJobId())
	entry.AccountId = strings.TrimSpace(entry.GetAccountId())
	entry.N8NExecutionId = strings.TrimSpace(entry.GetN8NExecutionId())
	entry.Action = strings.TrimSpace(entry.GetAction())
	entry.StepName = strings.TrimSpace(entry.GetStepName())
	entry.Channel = normalizeActionOTPChannel(entry.GetChannel())
	entry.Target = normalizeChannelOTPTarget(entry.GetChannel(), entry.GetTarget())
	entry.ResumeUrl = strings.TrimSpace(entry.GetResumeUrl())
	if entry.TimeoutSeconds <= 0 {
		entry.TimeoutSeconds = defaultChannelOTPTimeoutSeconds(entry.GetChannel())
	}
	if entry.CreatedAtUnix <= 0 {
		entry.CreatedAtUnix = time.Now().Unix()
	}
	return entry
}

func normalizeLatestChannelOTP(otp *pb.LatestChannelOTP) *pb.LatestChannelOTP {
	if otp == nil {
		otp = &pb.LatestChannelOTP{}
	}
	otp.Channel = normalizeActionOTPChannel(otp.GetChannel())
	otp.Target = normalizeChannelOTPTarget(otp.GetChannel(), otp.GetTarget())
	otp.Otp = normalizeOTP(otp.GetOtp())
	otp.Source = stringx.FirstNonEmpty(otp.GetSource(), otp.GetChannel())
	if otp.ReceivedAtUnix <= 0 {
		otp.ReceivedAtUnix = time.Now().Unix()
	}
	return otp
}

func channelOTPWaitAccepts(entry *pb.ChannelOTPWaitEntry, receivedAtUnix int64) bool {
	if entry == nil {
		return false
	}
	if entry.GetIssuedAfterUnix() > 0 && receivedAtUnix > 0 && receivedAtUnix < entry.GetIssuedAfterUnix() {
		return false
	}
	if entry.GetTimeoutSeconds() > 0 && entry.GetIssuedAfterUnix() > 0 && receivedAtUnix > entry.GetIssuedAfterUnix()+int64(entry.GetTimeoutSeconds()) {
		return false
	}
	return true
}

func channelOTPWaitTTL(timeoutSeconds int32) time.Duration {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 300
	}
	return time.Duration(timeoutSeconds)*time.Second + time.Hour
}

func defaultChannelOTPTimeoutSeconds(channel string) int32 {
	switch normalizeActionOTPChannel(channel) {
	case "sms", "wa":
		return 300
	default:
		return 300
	}
}

func channelOTPIndexKeys(channel string, target string) []string {
	channel = normalizeActionOTPChannel(channel)
	target = normalizeChannelOTPTarget(channel, target)
	if channel == "" || target == "" {
		return nil
	}
	candidates := []string{target}
	if channel == "wa" || channel == "sms" {
		if number := normalizeChannelOTPNumber(target); number != "" {
			candidates = append(candidates, number, strings.TrimPrefix(number, "+"))
		}
	}
	out := make([]string, 0, len(candidates))
	for _, candidate := range cleanStrings(candidates...) {
		out = append(out, channel+":"+candidate)
	}
	return out
}

func channelOTPAccountIndexKey(accountID string) string {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return ""
	}
	return "channel-otp-wait:account:" + accountID
}

func normalizeChannelOTPTarget(channel string, target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return ""
	}
	switch normalizeActionOTPChannel(channel) {
	case "wa", "sms":
		if number := normalizeChannelOTPNumber(target); number != "" {
			return number
		}
	}
	return target
}

func normalizeChannelOTPNumber(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var builder strings.Builder
	hasDigit := false
	for i, r := range value {
		if i == 0 && r == '+' {
			builder.WriteRune(r)
			continue
		}
		if unicode.IsDigit(r) {
			hasDigit = true
			builder.WriteRune(r)
			continue
		}
		if unicode.IsSpace(r) || r == '-' || r == '(' || r == ')' || r == '.' {
			continue
		}
		return ""
	}
	if !hasDigit {
		return ""
	}
	return strings.TrimSpace(builder.String())
}

func cleanStrings(values ...string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
