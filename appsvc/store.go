package appsvc

import (
	"context"
	"fmt"
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
