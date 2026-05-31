package appsvc

import (
	"strings"
	"time"
	"unicode"
)

func normalizeChannelOTPWaitEntry(entry channelOTPWaitEntry) channelOTPWaitEntry {
	entry.JobID = strings.TrimSpace(entry.JobID)
	entry.AccountID = strings.TrimSpace(entry.AccountID)
	entry.N8NExecutionID = strings.TrimSpace(entry.N8NExecutionID)
	entry.Action = strings.TrimSpace(entry.Action)
	entry.StepName = strings.TrimSpace(entry.StepName)
	entry.Channel = normalizeActionOTPChannel(entry.Channel)
	entry.Target = normalizeChannelOTPTarget(entry.Channel, entry.Target)
	entry.ResumeURL = strings.TrimSpace(entry.ResumeURL)
	if entry.TimeoutSeconds <= 0 {
		entry.TimeoutSeconds = defaultChannelOTPTimeoutSeconds(entry.Channel)
	}
	if entry.CreatedAtUnix <= 0 {
		entry.CreatedAtUnix = time.Now().Unix()
	}
	return entry
}

func normalizeLatestChannelOTP(otp latestChannelOTP) latestChannelOTP {
	otp.Channel = normalizeActionOTPChannel(otp.Channel)
	otp.Target = normalizeChannelOTPTarget(otp.Channel, otp.Target)
	otp.OTP = normalizeOTP(otp.OTP)
	otp.Source = firstNonEmpty(otp.Source, otp.Channel)
	if otp.ReceivedAtUnix <= 0 {
		otp.ReceivedAtUnix = time.Now().Unix()
	}
	return otp
}

func channelOTPWaitAccepts(entry channelOTPWaitEntry, receivedAtUnix int64) bool {
	if entry.IssuedAfterUnix > 0 && receivedAtUnix > 0 && receivedAtUnix < entry.IssuedAfterUnix {
		return false
	}
	if entry.TimeoutSeconds > 0 && entry.IssuedAfterUnix > 0 && receivedAtUnix > entry.IssuedAfterUnix+int64(entry.TimeoutSeconds) {
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
