package appsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OTPWebhookServer struct {
	httpServer *http.Server
	listener   net.Listener
}

type otpWebhookPayload struct {
	OTP string `json:"otp"`
}

type otpWebhookResponse struct {
	Success      bool   `json:"success"`
	Purpose      string `json:"purpose,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type otpSubmitPayload struct {
	Channel string `json:"channel"`
	Target  string `json:"target"`
	OTP     string `json:"otp"`
}

func StartOTPWebhook(addr string, submitURL string) (*OTPWebhookServer, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil, nil
	}
	submitURL = strings.TrimSpace(submitURL)
	if submitURL == "" {
		return nil, errors.New("gopay otp submit url is required")
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen gopay otp webhook %s: %w", addr, err)
	}
	handler := otpWebhookHandler{submitURL: submitURL, client: &http.Client{Timeout: 15 * time.Second}}
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	out := &OTPWebhookServer{httpServer: server, listener: listener}
	go func() {
		log.Printf("[gopay-app] OTP webhook listening on %s", listener.Addr().String())
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("[gopay-app] OTP webhook stopped: %v", err)
		}
	}()
	return out, nil
}

func (s *OTPWebhookServer) Close() error {
	if s == nil || s.httpServer == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

type otpWebhookHandler struct {
	submitURL string
	client    *http.Client
}

func (h otpWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/health" {
		writeOTPWebhookJSON(w, http.StatusOK, otpWebhookResponse{Success: true})
		return
	}
	if r.Method != http.MethodPost {
		writeOTPWebhookJSON(w, http.StatusMethodNotAllowed, otpWebhookResponse{Success: false, ErrorMessage: "method not allowed"})
		return
	}
	target, purpose, err := parseOTPWebhookPath(r.URL.Path)
	if err != nil {
		writeOTPWebhookJSON(w, http.StatusBadRequest, otpWebhookResponse{Success: false, ErrorMessage: err.Error()})
		return
	}
	var payload otpWebhookPayload
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024))
	if err := decoder.Decode(&payload); err != nil {
		writeOTPWebhookJSON(w, http.StatusBadRequest, otpWebhookResponse{Success: false, ErrorMessage: "invalid json payload"})
		return
	}
	code := normalizeOTP(payload.OTP)
	if code == "" {
		writeOTPWebhookJSON(w, http.StatusBadRequest, otpWebhookResponse{Success: false, ErrorMessage: "otp is required"})
		return
	}
	if err := h.postSubmit(r.Context(), otpSubmitPayload{Channel: "wa", Target: target, OTP: code}); err != nil {
		writeOTPWebhookJSON(w, http.StatusBadGateway, otpWebhookResponse{Success: false, ErrorMessage: err.Error()})
		return
	}
	log.Printf("[gopay-app] OTP webhook accepted purpose=%s target=%s", purpose, target)
	writeOTPWebhookJSON(w, http.StatusAccepted, otpWebhookResponse{Success: true, Purpose: target + "/" + purpose})
}

func (h otpWebhookHandler) postSubmit(ctx context.Context, payload otpSubmitPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.submitURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := h.client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("post gopay otp submit: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("gopay otp submit returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func parseOTPWebhookPath(path string) (string, string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("webhook path must be /<target>/<purpose>")
	}
	target, err := url.PathUnescape(parts[0])
	if err != nil {
		return "", "", fmt.Errorf("invalid target path segment")
	}
	purpose, err := url.PathUnescape(parts[1])
	if err != nil {
		return "", "", fmt.Errorf("invalid purpose path segment")
	}
	target = strings.TrimSpace(target)
	purpose = strings.ToLower(strings.TrimSpace(purpose))
	if target == "" || purpose == "" || strings.Contains(purpose, "/") {
		return "", "", fmt.Errorf("invalid otp route")
	}
	return target, purpose, nil
}

func writeOTPWebhookJSON(w http.ResponseWriter, status int, response otpWebhookResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func normalizeOTP(value string) string {
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "", "-", "")
	return strings.TrimSpace(replacer.Replace(value))
}
