package appsvc

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/byte-v-forge/common-lib/protojsonx"
	"github.com/byte-v-forge/gopay-app/pb"
	"github.com/redis/go-redis/v9"
)

const goPaySettingsKey = "settings:gopay"

var goPaySMSPricePattern = regexp.MustCompile(`^\d+(\.\d+)?$`)

func defaultGoPayRegisterIndonesiaWASettings() *pb.GoPayRegisterIndonesiaWASettings {
	return &pb.GoPayRegisterIndonesiaWASettings{
		SmsAcquireWaitSeconds:  90,
		SmsMinAvailableCount:   1,
		PhoneNumberMaxAttempts: 10,
	}
}

func normalizeGoPayRegisterIndonesiaWASettings(in *pb.GoPayRegisterIndonesiaWASettings) *pb.GoPayRegisterIndonesiaWASettings {
	out := defaultGoPayRegisterIndonesiaWASettings()
	if in == nil {
		return out
	}
	if value := in.GetSmsAcquireWaitSeconds(); value > 0 {
		out.SmsAcquireWaitSeconds = value
	}
	if value := in.GetSmsMinAvailableCount(); value > 0 {
		out.SmsMinAvailableCount = value
	}
	if value := in.GetPhoneNumberMaxAttempts(); value > 0 {
		out.PhoneNumberMaxAttempts = value
	}
	if value := strings.TrimSpace(in.GetSmsMinPriceAmountDecimal()); value != "" && goPaySMSPricePattern.MatchString(value) {
		out.SmsMinPriceAmountDecimal = value
	}
	if value := strings.TrimSpace(in.GetSmsMaxPriceAmountDecimal()); value != "" && goPaySMSPricePattern.MatchString(value) {
		out.SmsMaxPriceAmountDecimal = value
	}
	return out
}

func (s *Server) LoadGoPaySettings(ctx context.Context) (*pb.GoPayRegisterIndonesiaWASettings, error) {
	if s == nil || s.store == nil || s.store.client == nil {
		return nil, fmt.Errorf("gopay-app settings store is not configured")
	}
	raw, err := s.store.client.Get(ctx, s.store.redisKey(goPaySettingsKey)).Bytes()
	if err == redis.Nil {
		return defaultGoPayRegisterIndonesiaWASettings(), nil
	}
	if err != nil {
		return nil, err
	}
	var settings pb.GoPayRegisterIndonesiaWASettings
	if err := protojsonx.Unmarshal(raw, &settings); err != nil {
		return nil, err
	}
	return normalizeGoPayRegisterIndonesiaWASettings(&settings), nil
}

func (s *Server) SaveGoPaySettings(ctx context.Context, settings *pb.GoPayRegisterIndonesiaWASettings) (*pb.GoPayRegisterIndonesiaWASettings, error) {
	if s == nil || s.store == nil || s.store.client == nil {
		return nil, fmt.Errorf("gopay-app settings store is not configured")
	}
	normalized := normalizeGoPayRegisterIndonesiaWASettings(settings)
	data, err := protojsonx.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	if err := s.store.client.Set(ctx, s.store.redisKey(goPaySettingsKey), data, 0).Err(); err != nil {
		return nil, err
	}
	return normalized, nil
}
