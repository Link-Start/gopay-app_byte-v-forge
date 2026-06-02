package appsvc

import (
	"net/http"

	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/gopay-app/pb"
)

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
