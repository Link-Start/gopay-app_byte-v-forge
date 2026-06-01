package appsvc

import (
	"context"
	"strings"
	"time"
)

func (s *Server) releaseProxyRuntimeState(ctx context.Context, rawState string) error {
	state, err := parseState(rawState)
	if err != nil {
		return err
	}
	return s.releaseProxyRuntimeSession(ctx, state)
}

func (s *Server) releaseProxyRuntimeSession(ctx context.Context, state stateMap) error {
	accountID := stateString(state, "_proxy_runtime_account_id")
	if accountID == "" {
		accountID = proxyRuntimeAccountID(stateString(state, "_gopay_account_id"))
	}
	baseURL := proxyRuntimeAPIBase(s.cfg.ProxyRuntimeHTTPAddr)
	return s.releaseProxyRuntimeAccount(ctx, baseURL, accountID)
}

func (s *Server) releaseProxyRuntimeAccount(ctx context.Context, baseURL string, accountID string) error {
	accountID = strings.TrimSpace(accountID)
	if baseURL == "" || accountID == "" {
		return nil
	}
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	return postProxyRuntime(releaseCtx, baseURL+"/leases/release", map[string]any{"account_id": accountID}, nil, 5*time.Second)
}
