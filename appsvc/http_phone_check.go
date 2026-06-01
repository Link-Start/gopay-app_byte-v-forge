package appsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/byte-v-forge/gopay-app/pb"
)

type gopayPhoneCheckRequest struct {
	Phone             string `json:"phone"`
	CountryCode       string `json:"country_code"`
	StateJSON         string `json:"state_json"`
	ReleaseProxyState bool   `json:"release_proxy_state"`
}

type gopayPhoneCheckResponse struct {
	Success             bool   `json:"success"`
	Available           bool   `json:"available"`
	Phone               string `json:"phone"`
	CountryCode         string `json:"country_code"`
	Status              string `json:"status"`
	ErrorMessage        string `json:"error_message,omitempty"`
	ProxyHash           string `json:"proxy_hash,omitempty"`
	DeviceFingerprint   string `json:"device_fingerprint,omitempty"`
	GeneratedProxyState bool   `json:"generated_proxy_state"`
}

func (h gopayHTTPHandler) handlePhoneCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	req, phone, countryCode, err := readGopayPhoneCheckRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	state := strings.TrimSpace(req.StateJSON)
	generatedState := false
	var gen *pb.GenerateDeviceProxyResponse
	if state == "" {
		state, gen, err = h.generatePhoneCheckState(r.Context(), countryCode)
		generatedState = true
	}
	if state != "" && (generatedState || req.ReleaseProxyState) {
		defer func() { _ = h.service.releaseProxyRuntimeState(r.Context(), state) }()
	}
	if err != nil {
		writeJSON(w, http.StatusOK, phoneCheckErrorResponse(phone, countryCode, "proxy_unavailable", err, gen, generatedState))
		return
	}
	resp, err := h.service.CheckPhone(r.Context(), &pb.CheckPhoneRequest{Phone: phone, CountryCode: countryCode, StateJson: state})
	if err != nil {
		writeJSON(w, http.StatusOK, phoneCheckErrorResponse(phone, countryCode, "check_failed", err, gen, generatedState))
		return
	}
	writeJSON(w, http.StatusOK, gopayPhoneCheckResponse{
		Success:             resp.GetStatus() != "error" && resp.GetStatus() != "rate_limited",
		Available:           resp.GetAvailable(),
		Phone:               phone,
		CountryCode:         countryCode,
		Status:              firstNonEmpty(resp.GetStatus(), "error"),
		ErrorMessage:        resp.GetErrorMessage(),
		ProxyHash:           resp.GetProxyHash(),
		DeviceFingerprint:   resp.GetDeviceFingerprint(),
		GeneratedProxyState: generatedState,
	})
}

func readGopayPhoneCheckRequest(w http.ResponseWriter, r *http.Request) (gopayPhoneCheckRequest, string, string, error) {
	var req gopayPhoneCheckRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		return req, "", "", fmt.Errorf("request body must be json")
	}
	if strings.TrimSpace(req.CountryCode) == "" {
		return req, "", "", fmt.Errorf("country_code is required")
	}
	countryCode := phoneCountryCode(Config{}, req.CountryCode)
	if strings.HasPrefix(strings.TrimSpace(req.Phone), "+") && !strings.HasPrefix(digitsOnly(req.Phone), strings.TrimPrefix(countryCode, "+")) {
		return req, "", countryCode, fmt.Errorf("phone country code does not match country_code")
	}
	phone := strings.TrimLeft(normalizePhone(req.Phone, countryCode), "0")
	if phone == "" {
		return req, "", countryCode, fmt.Errorf("phone is required")
	}
	return req, phone, countryCode, nil
}

func (h gopayHTTPHandler) generatePhoneCheckState(ctx context.Context, countryCode string) (string, *pb.GenerateDeviceProxyResponse, error) {
	gen, err := h.service.GenerateDeviceProxy(ctx, &pb.GenerateDeviceProxyRequest{CountryCode: countryCode, ForceNew: true, SkipPreflight: true, EphemeralProfile: true})
	if err != nil {
		return "", nil, err
	}
	if !gen.GetSuccess() {
		return gen.GetStateJson(), gen, errors.New(firstNonEmpty(gen.GetErrorMessage(), "generate gopay device proxy failed"))
	}
	return gen.GetStateJson(), gen, nil
}

func phoneCheckErrorResponse(phone string, countryCode string, status string, err error, gen *pb.GenerateDeviceProxyResponse, generatedState bool) gopayPhoneCheckResponse {
	out := gopayPhoneCheckResponse{Success: false, Available: false, Phone: phone, CountryCode: countryCode, Status: status, ErrorMessage: err.Error(), GeneratedProxyState: generatedState}
	if gen != nil {
		out.ProxyHash = gen.GetProxyHash()
		out.DeviceFingerprint = gen.GetDeviceFingerprint()
	}
	return out
}
