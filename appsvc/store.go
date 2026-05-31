package appsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/accountcrud"
	"github.com/byte-v-forge/common-lib/accountstate"
	"github.com/byte-v-forge/common-lib/redisx"
	"github.com/redis/go-redis/v9"
)

type StateStore struct {
	clientClose func() error
	client      *redis.Client
	keyspace    redisx.Keyspace
	ttl         time.Duration
	accounts    *accountcrud.Manager[accountstate.AccountJSONRecord]
	profiles    *accountstate.JSONStore
}

func NewStateStore(ctx context.Context, redisURL string, keyPrefix string, ttl time.Duration) (*StateStore, error) {
	client, err := redisx.NewRequiredClient(ctx, redisURL, "GOPAY_STATE_REDIS_URL is required for gopay-app runtime state")
	if err != nil {
		return nil, err
	}
	accountStore := accountstate.NewAccountJSONStore(accountstate.AccountJSONStoreConfig{Client: client, Prefix: keyPrefix, TTL: ttl, Descriptor: gopayAccountDescriptor, IDField: "gopay_account_id"})
	return &StateStore{
		clientClose: client.Close,
		client:      client,
		keyspace:    redisx.NewKeyspace(keyPrefix),
		ttl:         ttl,
		accounts: accountcrud.New[accountstate.AccountJSONRecord](accountcrud.Config[accountstate.AccountJSONRecord]{
			Store:      accountcrud.NewAccountJSONStore(accountStore),
			Descriptor: gopayAccountDescriptor,
			IDField:    "gopay_account_id",
		}),
		profiles: accountstate.NewJSONStore(accountstate.JSONStoreConfig{Client: client, Prefix: keyPrefix, TTL: ttl}),
	}, nil
}

func NormalizeGopayAccountID(value string) (string, error) {
	return gopayAccountDescriptor.NormalizeID(value, "gopay_account_id")
}

func (s *StateStore) LoadAccount(ctx context.Context, gopayAccountID string) (string, error) {
	if s == nil || s.accounts == nil {
		return "", fmt.Errorf("gopay-app account state store is not configured")
	}
	record, found, err := s.accounts.Get(ctx, gopayAccountID)
	if err != nil {
		return "", err
	}
	if !found {
		return "{}", nil
	}
	return record.Raw, nil
}

func (s *StateStore) SaveAccount(ctx context.Context, gopayAccountID string, raw string) (string, error) {
	if s == nil || s.accounts == nil {
		return "", fmt.Errorf("gopay-app account state store is not configured")
	}
	record, err := s.accounts.Upsert(ctx, accountstate.AccountJSONRecord{AccountID: gopayAccountID, Raw: raw})
	if err != nil {
		return "", err
	}
	return record.Raw, nil
}

func (s *StateStore) DeleteAccount(ctx context.Context, gopayAccountID string) error {
	if s == nil || s.accounts == nil {
		return fmt.Errorf("gopay-app account state store is not configured")
	}
	_, err := s.accounts.Delete(ctx, gopayAccountID)
	return err
}

func (s *StateStore) ListAccounts(ctx context.Context, cursor string, limit int) (accountcrud.Page[accountstate.AccountJSONRecord], error) {
	if s == nil || s.accounts == nil {
		return accountcrud.Page[accountstate.AccountJSONRecord]{}, fmt.Errorf("gopay-app account state store is not configured")
	}
	return s.accounts.List(ctx, accountcrud.ListRequest{Cursor: cursor, Limit: limit})
}

func (s *StateStore) Load(ctx context.Context, key string) (string, error) {
	if s == nil || s.profiles == nil {
		return "", fmt.Errorf("gopay-app profile state store is not configured")
	}
	return s.profiles.LoadDefault(ctx, key, "{}")
}

func (s *StateStore) Save(ctx context.Context, key string, raw string) (string, error) {
	if s == nil || s.profiles == nil {
		return "", fmt.Errorf("gopay-app profile state store is not configured")
	}
	return s.profiles.Save(ctx, key, raw)
}

func (s *StateStore) Delete(ctx context.Context, key string) error {
	if s == nil || s.profiles == nil {
		return fmt.Errorf("gopay-app profile state store is not configured")
	}
	return s.profiles.Delete(ctx, key)
}

type channelOTPWaitEntry struct {
	JobID           string `json:"job_id"`
	AccountID       string `json:"account_id,omitempty"`
	N8NExecutionID  string `json:"n8n_execution_id,omitempty"`
	Action          string `json:"action,omitempty"`
	StepName        string `json:"step_name"`
	Channel         string `json:"channel"`
	Target          string `json:"target"`
	IssuedAfterUnix int64  `json:"issued_after_unix,omitempty"`
	TimeoutSeconds  int32  `json:"timeout_seconds,omitempty"`
	ResumeURL       string `json:"resume_url"`
	CreatedAtUnix   int64  `json:"created_at_unix"`
}

type latestChannelOTP struct {
	Channel        string `json:"channel"`
	Target         string `json:"target"`
	OTP            string `json:"otp"`
	ReceivedAtUnix int64  `json:"received_at_unix"`
	Source         string `json:"source"`
}

func (s *StateStore) RegisterChannelOTPWait(ctx context.Context, entry channelOTPWaitEntry, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("gopay-app channel otp wait store is not configured")
	}
	entry = normalizeChannelOTPWaitEntry(entry)
	if entry.JobID == "" || entry.Channel == "" || entry.Target == "" || entry.StepName == "" || entry.ResumeURL == "" {
		return fmt.Errorf("job_id, channel, target, step_name and resume_url are required")
	}
	if ttl <= 0 {
		ttl = s.ttl
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_ = s.DeleteChannelOTPWait(ctx, entry)
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, s.redisKey("channel-otp-wait:job:"+entry.JobID), data, ttl)
	for _, index := range channelOTPIndexKeys(entry.Channel, entry.Target) {
		key := s.redisKey("channel-otp-wait:index:" + index)
		pipe.SAdd(ctx, key, entry.JobID)
		pipe.Expire(ctx, key, ttl)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (s *StateStore) PendingChannelOTPWaits(ctx context.Context, channel string, target string, receivedAtUnix int64) ([]channelOTPWaitEntry, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("gopay-app channel otp wait store is not configured")
	}
	seen := map[string]struct{}{}
	out := []channelOTPWaitEntry{}
	for _, index := range channelOTPIndexKeys(channel, target) {
		ids, err := s.client.SMembers(ctx, s.redisKey("channel-otp-wait:index:"+index)).Result()
		if err != nil {
			return nil, err
		}
		for _, id := range cleanStrings(ids...) {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			entry, found, err := s.LoadChannelOTPWait(ctx, id)
			if err != nil {
				return nil, err
			}
			if !found {
				continue
			}
			if !channelOTPWaitAccepts(entry, receivedAtUnix) {
				continue
			}
			out = append(out, entry)
		}
	}
	return out, nil
}

func (s *StateStore) LoadChannelOTPWait(ctx context.Context, jobID string) (channelOTPWaitEntry, bool, error) {
	if s == nil || s.client == nil {
		return channelOTPWaitEntry{}, false, fmt.Errorf("gopay-app channel otp wait store is not configured")
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return channelOTPWaitEntry{}, false, nil
	}
	raw, err := s.client.Get(ctx, s.redisKey("channel-otp-wait:job:"+jobID)).Bytes()
	if err == redis.Nil {
		return channelOTPWaitEntry{}, false, nil
	}
	if err != nil {
		return channelOTPWaitEntry{}, false, err
	}
	var entry channelOTPWaitEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return channelOTPWaitEntry{}, false, err
	}
	return normalizeChannelOTPWaitEntry(entry), true, nil
}

func (s *StateStore) DeleteChannelOTPWait(ctx context.Context, entry channelOTPWaitEntry) error {
	if s == nil || s.client == nil {
		return nil
	}
	entry = normalizeChannelOTPWaitEntry(entry)
	if entry.JobID == "" {
		return nil
	}
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.redisKey("channel-otp-wait:job:"+entry.JobID), s.redisKey("channel-otp-wait:claim:"+entry.JobID))
	for _, index := range channelOTPIndexKeys(entry.Channel, entry.Target) {
		pipe.SRem(ctx, s.redisKey("channel-otp-wait:index:"+index), entry.JobID)
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

func (s *StateStore) SaveLatestChannelOTP(ctx context.Context, otp latestChannelOTP, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("gopay-app channel otp latest store is not configured")
	}
	otp = normalizeLatestChannelOTP(otp)
	if otp.Channel == "" || otp.Target == "" || otp.OTP == "" {
		return fmt.Errorf("channel, target and otp are required")
	}
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	data, err := json.Marshal(otp)
	if err != nil {
		return err
	}
	for _, index := range channelOTPIndexKeys(otp.Channel, otp.Target) {
		if err := s.client.Set(ctx, s.redisKey("channel-otp-latest:"+index), data, ttl).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *StateStore) LatestChannelOTP(ctx context.Context, channel string, target string, issuedAfterUnix int64, timeoutSeconds int32) (latestChannelOTP, bool, error) {
	if s == nil || s.client == nil {
		return latestChannelOTP{}, false, fmt.Errorf("gopay-app channel otp latest store is not configured")
	}
	for _, index := range channelOTPIndexKeys(channel, target) {
		raw, err := s.client.Get(ctx, s.redisKey("channel-otp-latest:"+index)).Bytes()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return latestChannelOTP{}, false, err
		}
		var otp latestChannelOTP
		if err := json.Unmarshal(raw, &otp); err != nil {
			return latestChannelOTP{}, false, err
		}
		otp = normalizeLatestChannelOTP(otp)
		if otp.OTP == "" {
			continue
		}
		if issuedAfterUnix > 0 && otp.ReceivedAtUnix > 0 && otp.ReceivedAtUnix < issuedAfterUnix {
			continue
		}
		if timeoutSeconds > 0 && issuedAfterUnix > 0 && otp.ReceivedAtUnix > issuedAfterUnix+int64(timeoutSeconds) {
			continue
		}
		return otp, true, nil
	}
	return latestChannelOTP{}, false, nil
}

func (s *StateStore) Close() error {
	if s == nil || s.clientClose == nil {
		return nil
	}
	return s.clientClose()
}

func (s *StateStore) redisKey(key string) string {
	redisKey, _ := s.keyspace.Key(key)
	return redisKey
}
