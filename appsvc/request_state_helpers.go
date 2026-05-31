package appsvc

import (
	"encoding/base64"
	"strings"
)

func (s *Server) parseRequestState(raw string) stateMap {
	state, err := parseState(raw)
	if err != nil {
		return stateMap{"last_error": err.Error()}
	}
	return state
}

func (s *Server) authBody(extra map[string]any) map[string]any {
	body := map[string]any{}
	for key, value := range extra {
		body[key] = value
	}
	body["client_id"] = s.cfg.GotoClientID
	body["client_secret"] = s.cfg.GotoClientSecret
	return body
}

func (s *Server) pin(value string) string {
	return strings.TrimSpace(value)
}

func (s *Server) signupProfile(phone, name, email string) (string, string) {
	resolvedName := strings.TrimSpace(name)
	resolvedEmail := strings.TrimSpace(email)
	if resolvedName != "" {
		return resolvedName, resolvedEmail
	}
	return signupNameFromSeed(signupSeed(phone)), resolvedEmail
}

func (s *Server) signupBasicAuthorization() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(s.cfg.SignupAuthUUID))
}
