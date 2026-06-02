package appsvc

import (
	"context"
	"net/http"
)

func (s *Server) getQrID(ctx context.Context, state stateMap) map[string]any {
	if stateString(state, "token") == "" {
		return map[string]any{"success": false, "error": "access_token missing"}
	}
	client, err := s.clientForState(ctx, state)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	resp, err := client.Customer.Get(ctx, "/v1/users/profile")
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if resp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": apiError("users/profile failed", resp)}
	}
	qrID := stringForAnyKey(resp.Data(), "qr_id", "qrId")
	if qrID == "" {
		return map[string]any{"success": false, "error": "qr_id not found in response: " + compactErrorDetail(resp.Payload)}
	}
	return map[string]any{"success": true, "qr_id": qrID}
}
