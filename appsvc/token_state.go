package appsvc

import (
	"time"

	"github.com/byte-v-forge/common-lib/jwtx"
)

func jwtExpiresAt(token string) int64 {
	return jwtx.ExpiresAt(token)
}

func tokenUsable(state stateMap, key string, minTTL time.Duration) bool {
	token := stateString(state, key)
	if token == "" {
		return false
	}
	expiresAt := jwtExpiresAt(token)
	if expiresAt == 0 {
		return true
	}
	return expiresAt > time.Now().Add(minTTL).Unix()
}
