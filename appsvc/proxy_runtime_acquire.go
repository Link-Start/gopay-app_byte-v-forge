package appsvc

import (
	"context"
	"fmt"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (s *Server) ensureProxyRuntimeSession(ctx context.Context, state stateMap, options proxyRuntimeAcquireOptions) error {
	if state == nil {
		return fmt.Errorf("gopay state missing")
	}
	identity := s.bindGoPayAccountIdentity(state, options.AccountID)
	if identity == "" {
		return fmt.Errorf("gopay account identity missing")
	}
	state["_gopay_country_code"] = normalizeGoPayProxyCountryCode(options.CountryCode)
	if !options.ForceNew && proxyRuntimeLeaseActive(state) && stateString(state, "_gopay_proxy") != "" && stateString(state, "_proxy_runtime_listener_id") != "" {
		return nil
	}
	sessionData, err := s.createProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{AccountID: identity, CountryCode: options.CountryCode, ForceNew: options.ForceNew, SkipPreflight: options.SkipPreflight, LeaseTTL: options.LeaseTTL})
	if err != nil {
		return err
	}
	for key, value := range sessionData {
		state[key] = value
	}
	return nil
}

func (s *Server) createProxyRuntimeSession(ctx context.Context, state stateMap, options proxyRuntimeAcquireOptions) (map[string]any, error) {
	baseURL := proxyRuntimeAPIBase(s.cfg.ProxyRuntimeHTTPAddr)
	if baseURL == "" {
		return nil, fmt.Errorf("PROXY_RUNTIME_HTTP_ADDR is required for GoPay dynamic IP proxy")
	}
	identity := s.bindGoPayAccountIdentity(state, options.AccountID)
	if identity == "" {
		return nil, fmt.Errorf("gopay account identity missing")
	}
	countryCode := normalizeGoPayProxyCountryCode(options.CountryCode)
	leaseTTL := stringx.FirstNonEmpty(options.LeaseTTL, goPayProxyLeaseTTL)
	if options.SkipPreflight {
		return s.acquireProxyRuntimeSession(ctx, baseURL, state, identity, countryCode, options.ForceNew, leaseTTL, 1, true)
	}
	var lastErr error
	for attempt := 1; attempt <= goPayProxyPreflightMaxAttempts; attempt++ {
		attemptData, err := s.acquireAndPreflightProxyRuntimeSession(ctx, baseURL, state, identity, countryCode, options.ForceNew || attempt > 1, leaseTTL, attempt)
		if err == nil {
			attemptData["_proxy_runtime_preflight_attempts"] = attempt
			return attemptData, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("gopay dynamic IP preflight failed after %d attempts: %w", goPayProxyPreflightMaxAttempts, lastErr)
}
