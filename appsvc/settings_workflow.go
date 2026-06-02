package appsvc

import (
	"context"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (h gopayHTTPHandler) registerIndonesiaWASettings(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	settings, err := h.service.LoadGoPaySettings(ctx)
	if err != nil {
		return nil, err
	}
	return h.baseResult(req, "register_indonesia_wa_settings", true, map[string]any{
		"sms_acquire_wait_seconds":     settings.GetSmsAcquireWaitSeconds(),
		"sms_min_available_count":      settings.GetSmsMinAvailableCount(),
		"sms_min_price_amount_decimal": settings.GetSmsMinPriceAmountDecimal(),
		"sms_max_price_amount_decimal": settings.GetSmsMaxPriceAmountDecimal(),
		"phone_number_max_attempts":    settings.GetPhoneNumberMaxAttempts(),
	}), nil
}

func (h gopayHTTPHandler) generateSharedPhoneCheckProxy(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	countryCode := stringx.FirstNonEmpty(req.CountryCode, mapString(req.Data, "country_code"), mapString(req.Data, "country_calling_code"), "62")
	requireLineProxy := false
	if normalizeGoPayWorkflowOperation(req.Operation) == registerIndonesiaWAWorkflowOperation {
		countryCode = stringx.FirstNonEmpty(mapString(req.Data, "proxy_country_code"), "US")
		requireLineProxy = true
	}
	state, err := h.service.generateDeviceProxyStateWithOptions(ctx, "", countryCode, true, true, true, goPayProxyLeaseTTL, requireLineProxy)
	rawState := stateJSON(state)
	_ = h.service.saveSharedPhoneCheckProxyState(ctx, req.JobID, rawState)
	data := h.service.deviceProxyDiagnostics(state)
	result := h.baseResult(req, "generate_shared_phone_check_proxy", err == nil, data)
	result.StateJSON = rawState
	if err != nil {
		result.ErrorMessage = err.Error()
	}
	return result, nil
}

func (h gopayHTTPHandler) releaseSharedPhoneCheckProxy(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	stateJSON := stringx.FirstNonEmpty(req.StateJSON, mapString(req.Data, "state_json"), mapString(req.Data, "proxy_state_json"))
	if stateJSON == "" {
		stateJSON = h.service.loadSharedPhoneCheckProxyState(ctx, req.JobID)
	}
	if stateJSON == "" {
		return h.baseResult(req, "release_shared_phone_check_proxy", true, nil), nil
	}
	if err := h.service.releaseProxyRuntimeState(ctx, stateJSON); err != nil {
		return nil, err
	}
	_ = h.service.deleteSharedPhoneCheckProxyState(ctx, req.JobID)
	return h.baseResult(req, "release_shared_phone_check_proxy", true, nil), nil
}

func sharedPhoneCheckProxyKey(jobID string) string {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return ""
	}
	return "shared-phone-check-proxy:" + jobID
}

func (s *Server) saveSharedPhoneCheckProxyState(ctx context.Context, jobID string, rawState string) error {
	key := sharedPhoneCheckProxyKey(jobID)
	if s == nil || s.store == nil || s.store.client == nil || key == "" || strings.TrimSpace(rawState) == "" || strings.TrimSpace(rawState) == "{}" {
		return nil
	}
	return s.store.client.Set(ctx, s.store.redisKey(key), rawState, 10*time.Minute).Err()
}

func (s *Server) loadSharedPhoneCheckProxyState(ctx context.Context, jobID string) string {
	key := sharedPhoneCheckProxyKey(jobID)
	if s == nil || s.store == nil || s.store.client == nil || key == "" {
		return ""
	}
	raw, err := s.store.client.Get(ctx, s.store.redisKey(key)).Result()
	if err != nil {
		return ""
	}
	return raw
}

func (s *Server) deleteSharedPhoneCheckProxyState(ctx context.Context, jobID string) error {
	key := sharedPhoneCheckProxyKey(jobID)
	if s == nil || s.store == nil || s.store.client == nil || key == "" {
		return nil
	}
	return s.store.client.Del(ctx, s.store.redisKey(key)).Err()
}
