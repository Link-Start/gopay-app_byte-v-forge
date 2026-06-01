package appsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/httpx"
	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/gopay-app/pb"
	"google.golang.org/protobuf/proto"
)

type HTTPAPIServer struct {
	httpServer *http.Server
	listener   net.Listener
}

func StartHTTPAPI(addr string, staticDir string, n8nWebhookBaseURL string, service *Server) (*HTTPAPIServer, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil, nil
	}
	if service == nil {
		return nil, fmt.Errorf("gopay-app service is required")
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen gopay HTTP API %s: %w", addr, err)
	}
	handler := gopayHTTPHandler{service: service, n8nWebhookBaseURL: strings.TrimRight(strings.TrimSpace(n8nWebhookBaseURL), "/")}
	mux := http.NewServeMux()
	mux.Handle("/api/gopay", handler)
	mux.Handle("/api/gopay/", handler)
	mux.Handle("/mf/gopay/", http.StripPrefix("/mf/gopay/", noCacheFileServer(staticDir)))
	mux.HandleFunc("/healthz", handler.handleHealth)
	server := &http.Server{Handler: withCORS(mux), ReadHeaderTimeout: 5 * time.Second}
	out := &HTTPAPIServer{httpServer: server, listener: listener}
	go func() {
		log.Printf("[gopay-app] HTTP API listening on %s", listener.Addr().String())
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("[gopay-app] HTTP API stopped: %v", err)
		}
	}()
	return out, nil
}

func (s *HTTPAPIServer) Close() error {
	if s == nil || s.httpServer == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

type gopayHTTPHandler struct {
	service           *Server
	n8nWebhookBaseURL string
}

func (h gopayHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/gopay"), "/")
	switch {
	case path == "" || path == "health":
		h.handleHealth(w, r)
	case path == "action-catalog":
		h.handleActionCatalog(w, r)
	case path == "accounts":
		h.handleAccounts(w, r)
	case strings.HasPrefix(path, "accounts/"):
		h.handleAccount(w, r, strings.TrimPrefix(path, "accounts/"))
	case path == "profile":
		h.handleProfile(w, r)
	case path == "otp/submit":
		h.handleOTPSubmit(w, r)
	case path == "phone/check":
		h.handlePhoneCheck(w, r)
	case path == "settings":
		h.handleSettings(w, r)
	case strings.HasPrefix(path, "workflows/"):
		h.handleWorkflowStart(w, r, strings.Trim(strings.TrimPrefix(path, "workflows/"), "/"))
	case strings.HasPrefix(path, "actions/gopay-account/"):
		h.handleAccountAction(w, r, strings.Trim(strings.TrimPrefix(path, "actions/gopay-account/"), "/"))
	case strings.HasPrefix(path, "actions/gopay-toolbox/"):
		h.handleToolboxAction(w, r, strings.Trim(strings.TrimPrefix(path, "actions/gopay-toolbox/"), "/"))
	case strings.HasPrefix(path, "actions/gopay-payment/"):
		h.handlePaymentAction(w, r, gopayPaymentActionScope, strings.Trim(strings.TrimPrefix(path, "actions/gopay-payment/"), "/"))
	default:
		writeError(w, http.StatusNotFound, fmt.Errorf("unknown gopay API path: %s", path))
	}
}

func (h gopayHTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success":                 true,
		"ok":                      true,
		"service":                 "gopay-app",
		"n8n_webhook_configured":  h.n8nWebhookBaseURL != "",
		"gopay_action_api_owned":  true,
		"gopay_account_api_owned": true,
		"workflows": []map[string]string{
			{"key": "gopay-account", "label": "GoPay 账户编排", "webhook_path": gopayAccountWebhookPath},
			{"key": gopayRegisterIndonesiaWAWorkflowKey, "label": registerIndonesiaWAWorkflowDisplayLabel, "webhook_path": gopayRegisterIndonesiaWAWebhookPath},
		},
	})
}

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

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,PUT,DELETE,OPTIONS")
		next.ServeHTTP(w, r)
	})
}

func noCacheFileServer(dir string) http.Handler {
	dir = strings.TrimSpace(dir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if dir == "" {
			http.NotFound(w, r)
			return
		}
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}
		http.NotFound(w, r)
	})
}

func writeProtoOrError(w http.ResponseWriter, value proto.Message, err error) {
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	_ = protojsonhttp.WriteResponse(w, http.StatusOK, value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
