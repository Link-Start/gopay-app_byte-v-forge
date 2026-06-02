package appsvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/protojsonx"
	"github.com/byte-v-forge/gopay-app/pb"
	"github.com/redis/go-redis/v9"
)

func (s *StateStore) LoadChannelOTPWait(ctx context.Context, jobID string) (*pb.ChannelOTPWaitEntry, bool, error) {
	if s == nil || s.client == nil {
		return nil, false, fmt.Errorf("gopay-app channel otp wait store is not configured")
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, false, nil
	}
	raw, err := s.client.Get(ctx, s.redisKey("channel-otp-wait:job:"+jobID)).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var entry pb.ChannelOTPWaitEntry
	if err := protojsonx.Unmarshal(raw, &entry); err != nil {
		return nil, false, err
	}
	return normalizeChannelOTPWaitEntry(&entry), true, nil
}

func (s *StateStore) DeleteChannelOTPWait(ctx context.Context, entry *pb.ChannelOTPWaitEntry) error {
	if s == nil || s.client == nil {
		return nil
	}
	entry = normalizeChannelOTPWaitEntry(entry)
	if entry.GetJobId() == "" {
		return nil
	}
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.redisKey("channel-otp-wait:job:"+entry.GetJobId()), s.redisKey("channel-otp-wait:claim:"+entry.GetJobId()))
	for _, index := range channelOTPIndexKeys(entry.GetChannel(), entry.GetTarget()) {
		pipe.SRem(ctx, s.redisKey("channel-otp-wait:index:"+index), entry.GetJobId())
	}
	if accountIndex := channelOTPAccountIndexKey(entry.GetAccountId()); accountIndex != "" {
		pipe.SRem(ctx, s.redisKey(accountIndex), entry.GetJobId())
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *StateStore) ClaimChannelOTPWait(ctx context.Context, jobID string, ttl time.Duration) (bool, error) {
	if s == nil || s.client == nil {
		return false, fmt.Errorf("gopay-app channel otp wait store is not configured")
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return false, nil
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	return s.client.SetNX(ctx, s.redisKey("channel-otp-wait:claim:"+jobID), "1", ttl).Result()
}

func (s *StateStore) ReleaseChannelOTPWaitClaim(ctx context.Context, jobID string) error {
	if s == nil || s.client == nil || strings.TrimSpace(jobID) == "" {
		return nil
	}
	return s.client.Del(ctx, s.redisKey("channel-otp-wait:claim:"+strings.TrimSpace(jobID))).Err()
}
