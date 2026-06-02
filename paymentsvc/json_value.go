package paymentsvc

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/byte-v-forge/common-lib/jsonx"
	"github.com/byte-v-forge/common-lib/jwtx"
	"github.com/byte-v-forge/common-lib/redactx"
)

func stringAt(value any, path ...string) string {
	return strings.TrimSpace(jsonx.StringAt(value, path...))
}

func boolAt(value any, path ...string) bool {
	return jsonx.BoolAt(value, path...)
}

func intAt(value any, path ...string) int64 {
	return jsonx.IntAt(value, path...)
}

func decodeJWTPayload(token string) map[string]any {
	return jwtx.PayloadOrNil(token)
}

func jsonExcerpt(value any, limit int) string {
	raw, err := jsonx.Compact(value)
	if err != nil {
		return redactx.Snippet(redactx.Text(fmt.Sprint(value)), limit)
	}
	return redactx.Snippet(redactx.Text(string(raw)), limit)
}

func mapValues(key, value string) url.Values {
	values := url.Values{}
	values.Set(key, value)
	return values
}
