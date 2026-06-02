package appsvc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/gopay-app/pb"
)

type OTPWebhookServer struct {
	httpServer *http.Server
	listener   net.Listener
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
		_ = protojsonhttp.WriteResponse(w, http.StatusOK, &pb.GopayOTPWebhookResponse{Success: true})
		return
	}
	if r.Method != http.MethodPost {
		_ = protojsonhttp.WriteResponse(w, http.StatusMethodNotAllowed, &pb.GopayOTPWebhookResponse{Success: false, ErrorMessage: "method not allowed"})
		return
	}
	target, purpose, err := parseOTPWebhookPath(r.URL.Path)
	if err != nil {
		_ = protojsonhttp.WriteResponse(w, http.StatusBadRequest, &pb.GopayOTPWebhookResponse{Success: false, ErrorMessage: err.Error()})
		return
	}
	var payload pb.GopayOTPWebhookRequest
	if err := protojsonhttp.ReadRequest(r, &payload); err != nil {
		_ = protojsonhttp.WriteResponse(w, http.StatusBadRequest, &pb.GopayOTPWebhookResponse{Success: false, ErrorMessage: "invalid json payload"})
		return
	}
	code := normalizeOTP(payload.GetOtp())
	if code == "" {
		_ = protojsonhttp.WriteResponse(w, http.StatusBadRequest, &pb.GopayOTPWebhookResponse{Success: false, ErrorMessage: "otp is required"})
		return
	}
	if err := h.postSubmit(r.Context(), &pb.SubmitChannelOTPRequest{Channel: "wa", Target: target, Otp: code}); err != nil {
		_ = protojsonhttp.WriteResponse(w, http.StatusBadGateway, &pb.GopayOTPWebhookResponse{Success: false, ErrorMessage: err.Error()})
		return
	}
	log.Printf("[gopay-app] OTP webhook accepted purpose=%s target=%s", purpose, target)
	_ = protojsonhttp.WriteResponse(w, http.StatusAccepted, &pb.GopayOTPWebhookResponse{Success: true, Purpose: target + "/" + purpose})
}

func (h otpWebhookHandler) postSubmit(ctx context.Context, payload *pb.SubmitChannelOTPRequest) error {
	opts := jsonPostOptions{Timeout: 15 * time.Second, Operation: "post gopay otp submit"}
	if h.client != nil {
		opts.Doer = h.client
	}
	if err := postJSON(ctx, h.submitURL, payload, opts); err != nil {
		return fmt.Errorf("post gopay otp submit: %w", err)
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

var otpCodeReplacer = strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "", "-", "")

func normalizeOTP(value string) string {
	return strings.TrimSpace(otpCodeReplacer.Replace(value))
}
