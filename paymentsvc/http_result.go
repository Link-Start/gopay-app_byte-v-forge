package paymentsvc

import (
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/redactx"
)

type httpResult struct {
	status  int
	headers stdhttp.Header
	body    []byte
	json    map[string]any
}

func (r *httpResult) data() map[string]any {
	if r == nil {
		return map[string]any{}
	}
	if data, ok := r.json["data"].(map[string]any); ok {
		return data
	}
	return r.json
}

func (r *httpResult) excerpt(limit int) string {
	if r == nil {
		return "<nil response>"
	}
	if limit <= 0 {
		limit = 600
	}
	text := strings.TrimSpace(string(r.body))
	if text == "" {
		raw, _ := json.Marshal(r.json)
		text = string(raw)
	}
	return redactx.Snippet(redactx.Text(text), limit)
}

func (r *httpResult) require(status int, label string) error {
	if r == nil {
		return fmt.Errorf("%s: empty response", label)
	}
	if r.status != status {
		return fmt.Errorf("%s %d: %s", label, r.status, r.excerpt(500))
	}
	return nil
}
