package appsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/byte-v-forge/common-lib/protojsonx"
	"github.com/byte-v-forge/gopay-app/pb"
	"github.com/redis/go-redis/v9"
)

func (s *StateStore) SaveLatestChannelOTP(ctx context.Context, otp *pb.LatestChannelOTP, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("gopay-app channel otp latest store is not configured")
	}
	otp = normalizeLatestChannelOTP(otp)
	if otp.GetChannel() == "" || otp.GetTarget() == "" || otp.GetOtp() == "" {
		return fmt.Errorf("channel, target and otp are required")
	}
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	data, err := protojsonx.Marshal(otp)
	if err != nil {
		return err
	}
	for _, index := range channelOTPIndexKeys(otp.GetChannel(), otp.GetTarget()) {
		if err := s.client.Set(ctx, s.redisKey("channel-otp-latest:"+index), data, ttl).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *StateStore) LatestChannelOTP(ctx context.Context, channel string, target string, issuedAfterUnix int64, timeoutSeconds int32) (*pb.LatestChannelOTP, bool, error) {
	if s == nil || s.client == nil {
		return nil, false, fmt.Errorf("gopay-app channel otp latest store is not configured")
	}
	for _, index := range channelOTPIndexKeys(channel, target) {
		raw, err := s.client.Get(ctx, s.redisKey("channel-otp-latest:"+index)).Bytes()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, false, err
		}
		var otp pb.LatestChannelOTP
		if err := protojsonx.Unmarshal(raw, &otp); err != nil {
			return nil, false, err
		}
		normalized := normalizeLatestChannelOTP(&otp)
		if normalized.GetOtp() == "" {
			continue
		}
		if issuedAfterUnix > 0 && normalized.GetReceivedAtUnix() > 0 && normalized.GetReceivedAtUnix() < issuedAfterUnix {
			continue
		}
		if timeoutSeconds > 0 && issuedAfterUnix > 0 && normalized.GetReceivedAtUnix() > issuedAfterUnix+int64(timeoutSeconds) {
			continue
		}
		return normalized, true, nil
	}
	return nil, false, nil
}
