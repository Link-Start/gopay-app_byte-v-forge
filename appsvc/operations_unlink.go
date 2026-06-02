package appsvc

import (
	"context"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/jsonx"
	"github.com/byte-v-forge/common-lib/stringx"
)

func (s *Server) unlink(ctx context.Context, state stateMap) map[string]any {
	client, err := s.clientForState(ctx, state)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	list, err := client.Customer.Get(ctx, "/v1/linkedapps")
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if list.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("linkedapps failed", list)}
	}
	services := linkedServices(list.Payload)
	unlinked := 0
	var failed []string
	for _, service := range services {
		path := jsonx.StringAt(service, "unlink_service_url")
		if path == "" {
			continue
		}
		name := stringx.FirstNonEmpty(jsonx.StringAt(service, "service_name"), jsonx.StringAt(service, "name"), path)
		resp, err := client.Customer.Patch(ctx, path, nil)
		if err == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusNoContent) {
			unlinked++
			continue
		}
		failed = append(failed, name)
	}
	if len(failed) > 0 {
		return map[string]any{"success": false, "error": "unlink failed: " + strings.Join(failed, ", "), "unlinked_count": unlinked}
	}
	state["stage"] = "consumed"
	delete(state, "last_error")
	return map[string]any{"success": true, "unlinked_count": unlinked}
}

func linkedServices(payload map[string]any) []map[string]any {
	var out []map[string]any
	var walk func(any)
	walk = func(value any) {
		switch typed := value.(type) {
		case map[string]any:
			if jsonx.StringAt(typed, "unlink_service_url") != "" {
				out = append(out, typed)
			}
			for _, child := range typed {
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(payload)
	return out
}
