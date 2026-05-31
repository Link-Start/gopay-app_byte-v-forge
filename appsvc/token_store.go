package appsvc

import (
	"time"

	"github.com/byte-v-forge/common-lib/jsonx"
)

func (s *Server) storeTokenResponse(state stateMap, data map[string]any, preserveRefresh bool) {
	token := jsonx.StringAt(data, "access_token")
	if token == "" {
		return
	}
	state["token"] = token
	refresh := jsonx.StringAt(data, "refresh_token")
	if refresh != "" {
		state["refresh_token"] = refresh
	} else if !preserveRefresh {
		delete(state, "refresh_token")
	}
	expiresAt := jwtExpiresAt(token)
	if expiresAt == 0 {
		expiresIn := anyInt(data["expires_in"])
		if expiresIn > 0 {
			expiresAt = time.Now().Unix() + expiresIn
		}
	}
	if expiresAt > 0 {
		state["token_expires_at"] = expiresAt
	} else {
		delete(state, "token_expires_at")
	}
	deleteKeys(state, "last_token_refresh_error", "last_token_refresh_failed_at")
}

func tmpTokenUsable(state stateMap, minTTL time.Duration) bool {
	token := stateString(state, "_tmp_token")
	if token == "" {
		return false
	}
	expiresAt := firstNonZero(jwtExpiresAt(token), stateInt(state, "_tmp_token_expires_at"))
	if expiresAt == 0 {
		return true
	}
	return expiresAt > time.Now().Add(minTTL).Unix()
}

func (s *Server) migrateActiveTokensToTmp(state stateMap, phone string) bool {
	moved := false
	for _, key := range activeTokenKeys {
		if value, ok := state[key]; ok && anyString(value) != "" {
			state["_tmp_"+key] = value
			moved = true
		}
		delete(state, key)
	}
	for _, key := range activeTokenMetaKeys {
		delete(state, key)
	}
	if moved {
		state["_tmp_token_migrated_at"] = time.Now().Unix()
		if phone != "" {
			state["_tmp_phone"] = phone
		}
	}
	return moved
}

func clearTmpTokens(state stateMap) {
	deleteKeys(state, tmpTokenKeys...)
	deleteKeys(state, tmpTokenMetaKeys...)
}
