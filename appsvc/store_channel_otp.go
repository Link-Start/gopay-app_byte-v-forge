package appsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/byte-v-forge/common-lib/protojsonx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (s *StateStore) RegisterChannelOTPWait(ctx context.Context, entry *pb.ChannelOTPWaitEntry, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("gopay-app channel otp wait store is not configured")
	}
	entry = normalizeChannelOTPWaitEntry(entry)
	if entry.GetJobId() == "" || entry.GetChannel() == "" || entry.GetTarget() == "" || entry.GetStepName() == "" || entry.GetResumeUrl() == "" {
		return fmt.Errorf("job_id, channel, target, step_name and resume_url are required")
	}
	if ttl <= 0 {
		ttl = s.ttl
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	data, err := protojsonx.Marshal(entry)
	if err != nil {
		return err
	}
	_ = s.DeleteChannelOTPWait(ctx, entry)
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, s.redisKey("channel-otp-wait:job:"+entry.GetJobId()), data, ttl)
	for _, index := range channelOTPIndexKeys(entry.GetChannel(), entry.GetTarget()) {
		key := s.redisKey("channel-otp-wait:index:" + index)
		pipe.SAdd(ctx, key, entry.GetJobId())
		pipe.Expire(ctx, key, ttl)
	}
	if accountIndex := channelOTPAccountIndexKey(entry.GetAccountId()); accountIndex != "" {
		key := s.redisKey(accountIndex)
		pipe.SAdd(ctx, key, entry.GetJobId())
		pipe.Expire(ctx, key, ttl)
	}
	_, err = pipe.Exec(ctx)
	return err
}
