package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/byte-v-forge/common-lib/jsonx"
)

func (s *Server) claimEnvelope(ctx context.Context, state stateMap, envelopeID, link string) map[string]any {
	resolved, err := resolveEnvelopeRequestID(envelopeID, link)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	client, err := s.clientForState(ctx, state)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	resp, err := client.Customer.Post(ctx, "/v1/festivals/envelope-requests", map[string]any{"envelope_request_id": resolved})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	raw := any(resp.Payload)
	data := resp.Data()
	if resp.StatusCode == http.StatusOK {
		detail, err := client.Customer.Get(ctx, "/v1/festivals/envelope-requests/"+url.PathEscape(resolved))
		if err == nil && detail.StatusCode == http.StatusOK {
			raw = map[string]any{"claim": resp.Payload, "detail": detail.Payload}
			data = detail.Data()
		}
	}
	success := resp.StatusCode == http.StatusOK
	errMessage := ""
	if !success {
		errMessage = apiError("claim envelope failed", resp)
	}
	return map[string]any{
		"success":                      success,
		"error":                        errMessage,
		"envelope_request_id":          resolved,
		"response_envelope_request_id": jsonx.StringAt(data, "envelope_request_id"),
		"status":                       jsonx.StringAt(data, "status"),
		"http_status":                  resp.StatusCode,
		"raw_json":                     safeJSON(raw),
	}
}

var (
	envelopeIDRE     = regexp.MustCompile(`[A-Za-z0-9_-]{8,128}`)
	envelopePathIDRE = regexp.MustCompile(`/v1/festivals/envelope-requests/([^/?#'" <]+)`)
)

func resolveEnvelopeRequestID(envelopeID, link string) (string, error) {
	for _, source := range []string{envelopeID, link} {
		for _, candidate := range decodedCandidates(source) {
			if match := envelopePathIDRE.FindStringSubmatch(candidate); len(match) > 1 {
				return match[1], nil
			}
			parsed, _ := url.Parse(candidate)
			if parsed != nil {
				query := parsed.Query()
				for _, key := range []string{"envelope_request_id", "envelopeRequestId", "id", "data", "link", "deep_link_value", "af_dp"} {
					for _, item := range query[key] {
						if resolved, err := resolveEnvelopeRequestID(item, ""); err == nil {
							return resolved, nil
						}
					}
				}
			}
			if envelopeIDRE.MatchString(candidate) && envelopeIDRE.FindString(candidate) == candidate {
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("envelope_request_id required, or pass a link containing one")
}

func decodedCandidates(value string) []string {
	current := strings.TrimSpace(value)
	var out []string
	seen := map[string]bool{}
	for range 5 {
		if current != "" && !seen[current] {
			out = append(out, current)
			seen[current] = true
		}
		decoded, err := url.QueryUnescape(current)
		if err != nil || decoded == current {
			break
		}
		current = decoded
	}
	return out
}
