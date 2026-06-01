package appsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/gopay-app/pb"
	"github.com/redis/go-redis/v9"
)

const goPaySettingsKey = "settings:gopay"

var goPaySMSPricePattern = regexp.MustCompile(`^\d+(\.\d+)?$`)

func defaultGoPayRegisterIndonesiaWASettings() *pb.GoPayRegisterIndonesiaWASettings {
	return &pb.GoPayRegisterIndonesiaWASettings{
		SmsAcquireWaitSeconds: 90,
		SmsMinAvailableCount:  1,
	}
}

func normalizeGoPayRegisterIndonesiaWASettings(in *pb.GoPayRegisterIndonesiaWASettings) *pb.GoPayRegisterIndonesiaWASettings {
	out := defaultGoPayRegisterIndonesiaWASettings()
	if in == nil {
		return out
	}
	if value := in.GetSmsAcquireWaitSeconds(); value > 0 {
		out.SmsAcquireWaitSeconds = value
	}
	if value := in.GetSmsMinAvailableCount(); value > 0 {
		out.SmsMinAvailableCount = value
	}
	if value := strings.TrimSpace(in.GetSmsMaxPriceAmountDecimal()); value != "" && goPaySMSPricePattern.MatchString(value) {
		out.SmsMaxPriceAmountDecimal = value
	}
	if out.GetSmsMaxPriceAmountDecimal() != "" {
		out.SmsMaxPriceCurrencyCode = strings.ToUpper(firstNonEmpty(in.GetSmsMaxPriceCurrencyCode(), "USD"))
	}
	return out
}

func (s *Server) LoadGoPaySettings(ctx context.Context) (*pb.GoPayRegisterIndonesiaWASettings, error) {
	if s == nil || s.store == nil || s.store.client == nil {
		return nil, fmt.Errorf("gopay-app settings store is not configured")
	}
	raw, err := s.store.client.Get(ctx, s.store.redisKey(goPaySettingsKey)).Bytes()
	if err == redis.Nil {
		return defaultGoPayRegisterIndonesiaWASettings(), nil
	}
	if err != nil {
		return nil, err
	}
	var settings pb.GoPayRegisterIndonesiaWASettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return nil, err
	}
	return normalizeGoPayRegisterIndonesiaWASettings(&settings), nil
}

func (s *Server) SaveGoPaySettings(ctx context.Context, settings *pb.GoPayRegisterIndonesiaWASettings) (*pb.GoPayRegisterIndonesiaWASettings, error) {
	if s == nil || s.store == nil || s.store.client == nil {
		return nil, fmt.Errorf("gopay-app settings store is not configured")
	}
	normalized := normalizeGoPayRegisterIndonesiaWASettings(settings)
	data, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	if err := s.store.client.Set(ctx, s.store.redisKey(goPaySettingsKey), data, 0).Err(); err != nil {
		return nil, err
	}
	return normalized, nil
}

func (h gopayHTTPHandler) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := h.service.LoadGoPaySettings(r.Context())
		if err != nil {
			writeProtoOrError(w, nil, err)
			return
		}
		_ = protojsonhttp.WriteResponse(w, http.StatusOK, &pb.GetGoPaySettingsResponse{Success: true, RegisterIndonesiaWa: settings})
	case http.MethodPost:
		var req pb.SaveGoPaySettingsRequest
		if err := protojsonhttp.ReadRequest(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		settings, err := h.service.SaveGoPaySettings(r.Context(), req.GetRegisterIndonesiaWa())
		if err != nil {
			writeProtoOrError(w, nil, err)
			return
		}
		_ = protojsonhttp.WriteResponse(w, http.StatusOK, &pb.SaveGoPaySettingsResponse{Success: true, RegisterIndonesiaWa: settings})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h gopayHTTPHandler) registerIndonesiaWASettings(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	settings, err := h.service.LoadGoPaySettings(ctx)
	if err != nil {
		return nil, err
	}
	return h.baseResult(req, "register_indonesia_wa_settings", true, map[string]any{
		"sms_acquire_wait_seconds":     settings.GetSmsAcquireWaitSeconds(),
		"sms_min_available_count":      settings.GetSmsMinAvailableCount(),
		"sms_max_price_amount_decimal": settings.GetSmsMaxPriceAmountDecimal(),
		"sms_max_price_currency_code":  settings.GetSmsMaxPriceCurrencyCode(),
	}), nil
}

func (h gopayHTTPHandler) generateSharedPhoneCheckProxy(ctx context.Context, req gopayActionRequest) (*gopayActionResult, error) {
	countryCode := firstNonEmpty(req.CountryCode, mapString(req.Data, "country_code"), mapString(req.Data, "country_calling_code"), "62")
	state, err := h.service.generateDeviceProxyStateWithLeaseTTL(ctx, "", countryCode, true, true, true, "180s")
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
	stateJSON := firstNonEmpty(req.StateJSON, mapString(req.Data, "state_json"), mapString(req.Data, "proxy_state_json"))
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
