package appsvc

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/httpx"
	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) handleActionCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	resp, err := h.service.GetActionCatalog(r.Context(), &pb.GetGopayActionCatalogRequest{})
	writeProtoOrError(w, resp, err)
}

func (h gopayHTTPHandler) handleAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.handleCreateAccount(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	resp, err := h.service.ListGopayAccounts(r.Context(), &pb.ListGopayAccountsRequest{
		Limit:  int32(httpx.QueryInt(r, "limit", 100)),
		Cursor: strings.TrimSpace(r.URL.Query().Get("cursor")),
	})
	writeProtoOrError(w, resp, err)
}

func (h gopayHTTPHandler) handleAccount(w http.ResponseWriter, r *http.Request, gopayAccountID string) {
	gopayAccountID = strings.Trim(gopayAccountID, "/")
	if gopayAccountID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("gopay_account_id is required"))
		return
	}
	switch r.Method {
	case http.MethodGet:
		resp, err := h.service.GetGopayAccount(r.Context(), &pb.GetGopayAccountRequest{GopayAccountId: gopayAccountID})
		writeProtoOrError(w, resp, err)
	case http.MethodDelete:
		resp, err := h.service.DeleteGopayAccount(r.Context(), &pb.DeleteGopayAccountRequest{GopayAccountId: gopayAccountID})
		writeProtoOrError(w, resp, err)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h gopayHTTPHandler) handleProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		gopayAccountID := strings.TrimSpace(r.URL.Query().Get("gopay_account_id"))
		resp, err := h.service.GetGopayAccountProfile(r.Context(), &pb.GetGopayAccountProfileRequest{GopayAccountId: gopayAccountID})
		writeProtoOrError(w, resp, err)
	case http.MethodPost:
		var req pb.SaveGopayAccountProfileRequest
		if err := protojsonhttp.ReadRequest(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		resp, err := h.service.SaveGopayAccountProfile(r.Context(), &req)
		writeProtoOrError(w, resp, err)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
