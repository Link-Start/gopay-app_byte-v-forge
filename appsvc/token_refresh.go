package appsvc

import (
	"context"
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/common-lib/jsonx"
	"github.com/byte-v-forge/common-lib/stringx"
)

func (s *Server) refreshAccessToken(ctx context.Context, state stateMap) map[string]any {
	refreshToken := stateString(state, "refresh_token")
	if refreshToken == "" {
		return map[string]any{"success": false, "error": "refresh_token missing"}
	}
	device, err := s.ensureDevice(state)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{}); err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	client, err := s.newClient(ctx, stateString(state, "token"), s.proxyForState(state), device)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	var last *httpjson.Response
	for _, body := range []map[string]any{
		s.authBody(map[string]any{"grant_type": "refresh_token", "token": refreshToken}),
		s.authBody(map[string]any{"grant_type": "refresh_token", "refresh_token": refreshToken}),
	} {
		resp, err := client.Auth.Post(ctx, "/goto-auth/token", body)
		if err != nil {
			state["last_token_refresh_error"] = err.Error()
			continue
		}
		last = resp
		if (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) && jsonx.StringAt(resp.Data(), "access_token") != "" {
			s.storeTokenResponse(state, resp.Data(), true)
			state["last_token_refresh_at"] = time.Now().Unix()
			deleteKeys(state, "last_token_refresh_error", "last_token_refresh_failed_at")
			if stateString(state, "last_error") == "TOKEN_REFRESH_FAILED" {
				delete(state, "last_error")
			}
			return map[string]any{"success": true, "refreshed": true, "expires_at": stateInt(state, "token_expires_at")}
		}
	}
	errMessage := apiError("refresh token failed", last)
	state["last_token_refresh_error"] = errMessage
	state["last_token_refresh_failed_at"] = time.Now().Unix()
	if !tokenUsable(state, "token", 0) {
		state["last_error"] = "TOKEN_REFRESH_FAILED"
	}
	return map[string]any{"success": false, "error": errMessage}
}

func (s *Server) ensureAccessToken(ctx context.Context, state stateMap, minTTL time.Duration, force bool) map[string]any {
	token := stateString(state, "token")
	expiresAt := jwtExpiresAt(token)
	if expiresAt > 0 {
		state["token_expires_at"] = expiresAt
	}
	if token != "" && !force && tokenUsable(state, "token", minTTL) {
		return map[string]any{"success": true, "refreshed": false, "expires_at": expiresAt}
	}
	result := s.refreshAccessToken(ctx, state)
	if anyBool(result["success"]) {
		return result
	}
	if token != "" && tokenUsable(state, "token", 0) {
		return map[string]any{"success": true, "refreshed": false, "expires_at": expiresAt, "warning": result["error"]}
	}
	return result
}

func (s *Server) verifyAccessToken(ctx context.Context, state stateMap) map[string]any {
	token := stateString(state, "token")
	if token == "" {
		return map[string]any{"success": false, "error": "access_token missing", "status": 0}
	}
	client, err := s.newClientWithState(ctx, state, false)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error(), "status": 0}
	}
	resp, err := client.Customer.Get(ctx, "/v1/users/profile")
	if err != nil {
		return map[string]any{"success": false, "error": err.Error(), "status": 0}
	}
	if resp.StatusCode == http.StatusOK {
		data := resp.Data()
		if profile := gojekCustomerProfile(data); len(profile) > 0 {
			s.syncProfileFields(state, profile, "")
		} else {
			phone := stringx.FirstNonEmpty(jsonx.StringAt(data, "phone"), jsonx.StringAt(data, "number"))
			if phone != "" {
				state["phone"] = normalizePhone(phone, "")
			}
		}
		pinSetup, pinSetupKnown := pinSetupFlagFromProfileData(data)
		if pinSetupKnown {
			updatePINSetupState(state, pinSetup)
		}
		state["stage"] = "ready"
		state["ready_at"] = time.Now().Unix()
		delete(state, "last_error")
		return map[string]any{"success": true, "status": 200, "phone": stateString(state, "phone"), "pin_setup": pinSetupKnown && pinSetup}
	}
	return map[string]any{"success": false, "status": resp.StatusCode, "error": apiError("profile failed", resp)}
}
