package appsvc

import (
	"context"
	"strings"
)

func (s *Server) resolveGoPayAccountPin(ctx context.Context, state stateMap, provided string) string {
	if pin := strings.TrimSpace(provided); pin != "" {
		return pin
	}
	if pin := strings.TrimSpace(stateString(state, "pin")); pin != "" {
		return pin
	}
	accountID := strings.TrimSpace(stateString(state, "_gopay_account_id"))
	if accountID == "" {
		return ""
	}
	profile, err := s.loadGopayAccountProfile(ctx, accountID)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(stateString(profile, "pin"))
}
