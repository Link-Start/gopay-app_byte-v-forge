package appsvc

import (
	"context"
	"net/http"
)

func (s *Server) checkPhoneByLoginMethods(ctx context.Context, phone, countryCode string, proxyState stateMap) map[string]any {
	cc := phoneCountryCode(s.cfg, countryCode)
	normalized := normalizePhoneWithConfig(s.cfg, phone, cc)
	if proxyState == nil {
		proxyState = stateMap{}
	}
	proxyURL := stateString(proxyState, "_gopay_proxy")
	if proxyURL == "" {
		return s.checkPhoneResult(proxyState, map[string]any{"success": false, "available": false, "status": "error", "error": "generated proxy missing", "attempts": 0})
	}
	rawDevice := nestedMap(proxyState["device"])
	if len(rawDevice) == 0 {
		return s.checkPhoneResult(proxyState, map[string]any{"success": false, "available": false, "status": "error", "error": "generated device missing", "attempts": 0})
	}
	device := deviceFromMap(rawDevice)
	if device.AppID == "" || device.UniqueID == "" || device.PhoneMake == "" || device.PhoneModel == "" {
		return s.checkPhoneResult(proxyState, map[string]any{"success": false, "available": false, "status": "error", "error": "generated device incomplete", "attempts": 0})
	}
	client, err := s.newClient(ctx, "", proxyURL, device)
	if err != nil {
		return s.checkPhoneResult(proxyState, map[string]any{"success": false, "available": false, "status": "error", "error": err.Error(), "attempts": 0})
	}
	resp, err := client.Auth.Post(ctx, "/goto-auth/login/methods", signupProbeBody{
		PhoneNumber:               normalized,
		CountryCode:               cc,
		Email:                     "",
		DeviceVerificationTokenID: "",
		ClientID:                  s.cfg.GotoClientID,
		ClientSecret:              s.cfg.GotoClientSecret,
	})
	if err != nil {
		return s.checkPhoneResult(proxyState, map[string]any{"success": false, "available": false, "status": "error", "error": err.Error(), "attempts": 1})
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		verificationID, methods, defaultMethod := s.persistLoginProbe(proxyState, normalized, cc, resp.Data())
		return s.checkPhoneResult(proxyState, map[string]any{
			"success": true, "available": false, "status": "registered",
			"verification_id_present": verificationID != "", "methods": methods, "default_method": defaultMethod,
			"attempts": 1,
		})
	}
	if loginMethodsInvalidUser(resp) {
		return s.checkPhoneResult(proxyState, map[string]any{"success": true, "available": true, "status": "available", "attempts": 1})
	}
	if isRateLimited(resp) {
		return s.checkPhoneResult(proxyState, map[string]any{"success": false, "available": false, "status": "rate_limited", "error": loginMethodsRateLimitedError(), "attempts": 1})
	}
	return s.checkPhoneResult(proxyState, map[string]any{"success": false, "available": false, "status": "error", "error": apiError("login methods failed", resp), "attempts": 1})
}

func (s *Server) checkPhoneResult(state stateMap, data map[string]any) map[string]any {
	if data == nil {
		data = map[string]any{}
	}
	for key, value := range s.deviceProxyDiagnostics(state) {
		data[key] = value
	}
	data["state_json"] = stateJSON(state)
	return data
}
