package appsvc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
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
