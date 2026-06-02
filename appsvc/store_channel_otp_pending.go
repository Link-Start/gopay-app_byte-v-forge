package appsvc

import (
	"context"
	"fmt"

	"github.com/byte-v-forge/gopay-app/pb"
)

func (s *StateStore) PendingChannelOTPWaits(ctx context.Context, channel string, target string, receivedAtUnix int64) ([]*pb.ChannelOTPWaitEntry, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("gopay-app channel otp wait store is not configured")
	}
	seen := map[string]struct{}{}
	out := []*pb.ChannelOTPWaitEntry{}
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

func (s *StateStore) PendingAccountOTPWaits(ctx context.Context, accountID string, receivedAtUnix int64) ([]*pb.ChannelOTPWaitEntry, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("gopay-app channel otp wait store is not configured")
	}
	accountIndex := channelOTPAccountIndexKey(accountID)
	if accountIndex == "" {
		return nil, nil
	}
	ids, err := s.client.SMembers(ctx, s.redisKey(accountIndex)).Result()
	if err != nil {
		return nil, err
	}
	out := []*pb.ChannelOTPWaitEntry{}
	for _, id := range cleanStrings(ids...) {
		entry, found, err := s.LoadChannelOTPWait(ctx, id)
		if err != nil {
			return nil, err
		}
		if !found || !channelOTPWaitAccepts(entry, receivedAtUnix) {
			continue
		}
		out = append(out, entry)
	}
	return out, nil
}
