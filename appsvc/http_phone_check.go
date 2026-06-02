package appsvc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) handlePhoneCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	req, phone, countryCode, err := readGopayPhoneCheckRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	state := strings.TrimSpace(req.GetStateJson())
	generatedState := false
	var gen *pb.GenerateDeviceProxyResponse
	if state == "" {
		state, gen, err = h.generatePhoneCheckState(r.Context(), countryCode)
		generatedState = true
	}
	if state != "" && (generatedState || req.GetReleaseProxyState()) {
		defer func() { _ = h.service.releaseProxyRuntimeState(r.Context(), state) }()
	}
	if err != nil {
		_ = protojsonhttp.WriteResponse(w, http.StatusOK, phoneCheckErrorResponse(phone, countryCode, "proxy_unavailable", err, gen, generatedState))
		return
	}
	resp, err := h.service.CheckPhone(r.Context(), &pb.CheckPhoneRequest{Phone: phone, CountryCode: countryCode, StateJson: state})
	if err != nil {
		_ = protojsonhttp.WriteResponse(w, http.StatusOK, phoneCheckErrorResponse(phone, countryCode, "check_failed", err, gen, generatedState))
		return
	}
	_ = protojsonhttp.WriteResponse(w, http.StatusOK, &pb.CheckGopayPhoneAvailabilityResponse{
		Success:             resp.GetStatus() != "error" && resp.GetStatus() != "rate_limited",
		Available:           resp.GetAvailable(),
		Phone:               phone,
		CountryCode:         countryCode,
		Status:              stringx.FirstNonEmpty(resp.GetStatus(), "error"),
		ErrorMessage:        resp.GetErrorMessage(),
		ProxyHash:           resp.GetProxyHash(),
		DeviceFingerprint:   resp.GetDeviceFingerprint(),
		GeneratedProxyState: generatedState,
	})
}

func readGopayPhoneCheckRequest(r *http.Request) (*pb.CheckGopayPhoneAvailabilityRequest, string, string, error) {
	var req pb.CheckGopayPhoneAvailabilityRequest
	if err := protojsonhttp.ReadRequest(r, &req); err != nil {
		return &req, "", "", fmt.Errorf("request body must be json")
	}
	if strings.TrimSpace(req.GetCountryCode()) == "" {
		return &req, "", "", fmt.Errorf("country_code is required")
	}
	countryCode := phoneCountryCode(Config{}, req.GetCountryCode())
	if strings.HasPrefix(strings.TrimSpace(req.GetPhone()), "+") && !strings.HasPrefix(digitsOnly(req.GetPhone()), strings.TrimPrefix(countryCode, "+")) {
		return &req, "", countryCode, fmt.Errorf("phone country code does not match country_code")
	}
	phone := strings.TrimLeft(normalizePhone(req.GetPhone(), countryCode), "0")
	if phone == "" {
		return &req, "", countryCode, fmt.Errorf("phone is required")
	}
	return &req, phone, countryCode, nil
}

func (h gopayHTTPHandler) generatePhoneCheckState(ctx context.Context, countryCode string) (string, *pb.GenerateDeviceProxyResponse, error) {
	gen, err := h.service.GenerateDeviceProxy(ctx, &pb.GenerateDeviceProxyRequest{CountryCode: countryCode, ForceNew: true, SkipPreflight: true, EphemeralProfile: true})
	if err != nil {
		return "", nil, err
	}
	if !gen.GetSuccess() {
		return gen.GetStateJson(), gen, errors.New(stringx.FirstNonEmpty(gen.GetErrorMessage(), "generate gopay device proxy failed"))
	}
	return gen.GetStateJson(), gen, nil
}

func phoneCheckErrorResponse(phone string, countryCode string, status string, err error, gen *pb.GenerateDeviceProxyResponse, generatedState bool) *pb.CheckGopayPhoneAvailabilityResponse {
	out := &pb.CheckGopayPhoneAvailabilityResponse{Success: false, Available: false, Phone: phone, CountryCode: countryCode, Status: status, ErrorMessage: err.Error(), GeneratedProxyState: generatedState}
	if gen != nil {
		out.ProxyHash = gen.GetProxyHash()
		out.DeviceFingerprint = gen.GetDeviceFingerprint()
	}
	return out
}
