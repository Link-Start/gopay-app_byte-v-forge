package appsvc

import (
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/jsonx"
)

type stateMap map[string]any

func parseState(raw string) (stateMap, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return stateMap{}, nil
	}
	if strings.EqualFold(raw, "null") {
		return stateMap{}, nil
	}
	if !strings.HasPrefix(raw, "{") {
		return nil, fmt.Errorf("state_json must be a JSON object")
	}
	value, err := jsonx.DecodeMap([]byte(raw))
	if err != nil {
		return nil, err
	}
	if value == nil {
		value = jsonx.Map{}
	}
	return stateMap(value), nil
}

func stateJSON(state stateMap) string {
	if state == nil {
		state = stateMap{}
	}
	raw, err := jsonx.Compact(state)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func stateString(state stateMap, key string) string {
	if state == nil {
		return ""
	}
	return anyString(state[key])
}

func stateInt(state stateMap, key string) int64 {
	if state == nil {
		return 0
	}
	return anyInt(state[key])
}

func deleteKeys(state stateMap, keys ...string) {
	for _, key := range keys {
		delete(state, key)
	}
}
