package appsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
	gopayapp "github.com/byte-v-forge/gopay-app/protocol/app"
)

func (s *Server) newClient(ctx context.Context, token string, proxyURL string, device gopayapp.DeviceFingerprint) (*gopayapp.ClientSet, error) {
	cfg := gopayapp.ConfigFromEnv(token)
	cfg.ProxyURL = proxyURL
	cfg.Timeout = 30 * time.Second
	cfg.Device = device
	cfg.Logger = func(ctx context.Context, message string, fields map[string]any) {
		fmt.Printf("[gopay-app] %s %v\n", message, fields)
	}
	return gopayapp.NewClientSet(cfg)
}

func (s *Server) clientForState(ctx context.Context, state stateMap) (*gopayapp.ClientSet, error) {
	refresh := s.ensureAccessToken(ctx, state, s.cfg.TokenRefreshMinTTL, false)
	if !anyBool(refresh["success"]) && !tokenUsable(state, "token", 0) {
		return nil, fmt.Errorf("%s", stringx.FirstNonEmpty(anyString(refresh["error"]), "token refresh failed"))
	}
	if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{}); err != nil {
		return nil, err
	}
	device, err := s.ensureDevice(state)
	if err != nil {
		return nil, err
	}
	return s.newClient(ctx, stateString(state, "token"), s.proxyForState(state), device)
}

func (s *Server) tmpClientForState(ctx context.Context, state stateMap) (*gopayapp.ClientSet, error) {
	token := stateString(state, "_tmp_token")
	if token == "" {
		return nil, fmt.Errorf("temporary account token missing")
	}
	if !tmpTokenUsable(state, 0) {
		expiresAt := firstNonZero(jwtExpiresAt(token), stateInt(state, "_tmp_token_expires_at"))
		return nil, fmt.Errorf("temporary account token expired: expires_at=%d", expiresAt)
	}
	if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{}); err != nil {
		return nil, err
	}
	device, err := s.ensureDevice(state)
	if err != nil {
		return nil, err
	}
	return s.newClient(ctx, token, s.proxyForState(state), device)
}

func (s *Server) rotateLoginAttemptIdentity(ctx context.Context, state stateMap) error {
	if state == nil {
		return nil
	}
	if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{ForceNew: true}); err != nil {
		return err
	}
	state["_proxy_runtime_session_rotated_for_login"] = true
	return nil
}

func (s *Server) proxyForState(state stateMap) string {
	return stateString(state, "_gopay_proxy")
}
