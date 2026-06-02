package appsvc

import (
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/jsonx"
)

func anyString(value any) string {
	return strings.TrimSpace(jsonx.String(value))
}

func anyInt(value any) int64 {
	return jsonx.Int(value)
}

func anyBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case int:
		return typed != 0
	case int32:
		return typed != 0
	case int64:
		return typed != 0
	case float32:
		return typed != 0
	case float64:
		return typed != 0
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "1", "true", "yes", "on":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func nestedMap(value any) map[string]any {
	if typed, ok := jsonx.Object(value); ok {
		return typed
	}
	return map[string]any{}
}

func safeJSON(value any) string {
	raw, err := jsonx.Compact(value)
	if err != nil {
		return fmt.Sprintf(`{"marshal_error":%q}`, err.Error())
	}
	return string(raw)
}

func compactErrorDetail(value any) string {
	raw := safeJSON(value)
	if len(raw) > 800 {
		return raw[:800]
	}
	return raw
}
