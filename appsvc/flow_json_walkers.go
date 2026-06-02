package appsvc

import (
	"strings"

	"github.com/byte-v-forge/common-lib/jsonx"
)

func intForAnyKey(value any, keys ...string) int64 {
	wanted := map[string]struct{}{}
	for _, key := range keys {
		wanted[jsonx.NormalizeKey(key)] = struct{}{}
	}
	var walk func(any) int64
	walk = func(current any) int64 {
		if obj, ok := jsonx.Object(current); ok {
			for key, item := range obj {
				if _, ok := wanted[jsonx.NormalizeKey(key)]; ok {
					if parsed := anyInt(item); parsed != 0 {
						return parsed
					}
				}
			}
			for _, item := range obj {
				if parsed := walk(item); parsed != 0 {
					return parsed
				}
			}
		}
		switch typed := current.(type) {
		case []any:
			for _, item := range typed {
				if parsed := walk(item); parsed != 0 {
					return parsed
				}
			}
		}
		return 0
	}
	return walk(value)
}

func boolForAnyKey(value any, keys ...string) bool {
	wanted := map[string]struct{}{}
	for _, key := range keys {
		wanted[jsonx.NormalizeKey(key)] = struct{}{}
	}
	var walk func(any) bool
	walk = func(current any) bool {
		if obj, ok := jsonx.Object(current); ok {
			for key, item := range obj {
				if _, ok := wanted[jsonx.NormalizeKey(key)]; ok {
					return anyBool(item)
				}
			}
			for _, item := range obj {
				if walk(item) {
					return true
				}
			}
		}
		if items, ok := current.([]any); ok {
			for _, item := range items {
				if walk(item) {
					return true
				}
			}
		}
		return false
	}
	return walk(value)
}

func stringForAnyKey(value any, keys ...string) string {
	return jsonx.StringAtAnyKey(value, keys...)
}

func verificationScopedID(value any) string {
	var walk func(any, bool) string
	walk = func(current any, inVerificationScope bool) string {
		if obj, ok := jsonx.Object(current); ok {
			for key, item := range obj {
				normalized := jsonx.NormalizeKey(key)
				nextScope := inVerificationScope || strings.Contains(normalized, "verification")
				if nextScope && normalized == "id" {
					if text := anyString(item); text != "" {
						return text
					}
				}
				if text := walk(item, nextScope); text != "" {
					return text
				}
			}
			return ""
		}
		switch typed := current.(type) {
		case []any:
			for _, item := range typed {
				if text := walk(item, inVerificationScope); text != "" {
					return text
				}
			}
		}
		return ""
	}
	return walk(value, false)
}
