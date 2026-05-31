package appsvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/hashx"
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
	sessionData, err := s.createProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{AccountID: identity, CountryCode: options.CountryCode, ForceNew: options.ForceNew, SkipPreflight: options.SkipPreflight})
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
	if options.SkipPreflight {
		return s.acquireProxyRuntimeSession(ctx, baseURL, state, identity, countryCode, options.ForceNew, 1, true)
	}
	var lastErr error
	for attempt := 1; attempt <= goPayProxyPreflightMaxAttempts; attempt++ {
		attemptData, err := s.acquireAndPreflightProxyRuntimeSession(ctx, baseURL, state, identity, countryCode, options.ForceNew || attempt > 1, attempt)
		if err == nil {
			attemptData["_proxy_runtime_preflight_attempts"] = attempt
			return attemptData, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("gopay dynamic IP preflight failed after %d attempts: %w", goPayProxyPreflightMaxAttempts, lastErr)
}

func (s *Server) acquireAndPreflightProxyRuntimeSession(ctx context.Context, baseURL string, state stateMap, identity string, countryCode string, forceNew bool, attempt int) (map[string]any, error) {
	out, listenerID, exitIP, err := s.acquireProxyRuntimeSessionWithListener(ctx, baseURL, state, identity, countryCode, forceNew, attempt, false)
	if err != nil {
		return nil, err
	}
	geo, err := s.checkProxyRuntimeGeo(ctx, baseURL, exitIP)
	if err != nil {
		return nil, err
	}
	if geo.ProxyExitGeo.CountryCode != "" && !strings.EqualFold(geo.ProxyExitGeo.CountryCode, countryCode) {
		return nil, fmt.Errorf("proxy exit country mismatch: got %s want %s", geo.ProxyExitGeo.CountryCode, countryCode)
	}
	fraud, err := s.checkProxyRuntimeFraud(ctx, baseURL, exitIP)
	if err != nil {
		return nil, err
	}
	if proxyIPFraudRiskRejected(fraud.Check.RiskLevel) {
		return nil, fmt.Errorf("proxy IP fraud risk rejected: level=%s score=%.0f", fraud.Check.RiskLevel, fraud.Check.RiskScore)
	}
	connectivity, err := s.checkGoPayProxyRuntimeConnectivity(ctx, baseURL, listenerID)
	if err != nil {
		return nil, err
	}
	out["_proxy_runtime_preflight_country_code"] = countryCode
	out["_proxy_runtime_preflight_attempt"] = attempt
	out["_proxy_runtime_exit_ip_hash"] = hashx.ShortSHA256(exitIP, 12)
	out["_proxy_runtime_exit_country_code"] = geo.ProxyExitGeo.CountryCode
	out["_proxy_runtime_exit_region"] = geo.ProxyExitGeo.Region
	out["_proxy_runtime_ip_fraud_risk_level"] = normalizeProxyRiskLevel(fraud.Check.RiskLevel)
	out["_proxy_runtime_ip_fraud_risk_score"] = fraud.Check.RiskScore
	out["_proxy_runtime_ip_network_kind"] = shortProxyEnum(fraud.Check.NetworkKind)
	out["_proxy_runtime_ip_anonymizer_kind"] = shortProxyEnum(fraud.Check.AnonymizerKind)
	out["_proxy_runtime_connectivity_reachable"] = true
	out["_proxy_runtime_connectivity_targets"] = connectivity
	return out, nil
}

func (s *Server) acquireProxyRuntimeSession(ctx context.Context, baseURL string, state stateMap, identity string, countryCode string, forceNew bool, attempt int, skipPreflight bool) (map[string]any, error) {
	out, _, _, err := s.acquireProxyRuntimeSessionWithListener(ctx, baseURL, state, identity, countryCode, forceNew, attempt, skipPreflight)
	return out, err
}

func (s *Server) acquireProxyRuntimeSessionWithListener(ctx context.Context, baseURL string, state stateMap, identity string, countryCode string, forceNew bool, attempt int, skipPreflight bool) (map[string]any, string, string, error) {
	accountID := proxyRuntimeAccountID(identity)
	leaseReq := map[string]any{
		"account_id": accountID,
		"purpose":    goPayProxyPurpose,
		"force_new":  forceNew,
		"policy": map[string]any{
			"mode":          "PROXY_SESSION_MODE_STICKY",
			"region":        countryCode,
			"sticky_ttl":    goPayProxyLeaseTTL,
			"upstream_kind": "PROXY_UPSTREAM_KIND_DYNAMIC_IP",
			"rotation_mode": "PROXY_ROTATION_MODE_STICKY_SESSION",
			"labels": map[string]string{
				"purpose": "gopay_app",
				"country": countryCode,
			},
		},
		"chain_policy": map[string]any{
			"country_code":                 countryCode,
			"purpose":                      goPayProxyPurpose,
			"strategy":                     "PROXY_CHAIN_STRATEGY_REGION_AWARE",
			"require_dynamic_exit":         true,
			"allow_direct_dynamic_gateway": true,
			"prefer_line_proxy":            true,
		},
	}
	var lease proxyRuntimeLeaseResponse
	if err := postProxyRuntime(ctx, baseURL+"/leases/acquire", leaseReq, &lease, 30*time.Second); err != nil {
		return nil, "", "", err
	}
	egress := lease.Egress
	if egress.Host == "" {
		egress = lease.Lease.Egress
	}
	proxyURL, err := proxyRuntimeProxyURL(egress)
	if err != nil {
		return nil, "", "", err
	}
	listenerID := strings.TrimSpace(lease.Lease.Listener.ListenerID)
	if listenerID == "" {
		return nil, "", "", fmt.Errorf("proxy-runtime returned lease without listener")
	}

	chainPlan := firstMap(lease.ChainPlan, lease.Lease.ChainPlan)
	out := map[string]any{
		"_gopay_proxy":                      proxyURL,
		"_gopay_account_id":                 identity,
		"_gopay_country_code":               countryCode,
		"_proxy_runtime_account_id":         accountID,
		"_proxy_runtime_lease_id":           lease.Lease.LeaseID,
		"_proxy_runtime_provider_account":   lease.Lease.ProviderAccountID,
		"_proxy_runtime_lease_expires_at":   lease.Lease.ExpiresAt,
		"_proxy_runtime_listener_id":        listenerID,
		"_proxy_runtime_listener_kind":      lease.Lease.Listener.Kind,
		"_proxy_runtime_session_started_at": time.Now().Unix(),
		"_proxy_runtime_pool_endpoints":     len(lease.Pool.Endpoints),
		"_proxy_runtime_session_rotated":    forceNew,
		"_proxy_runtime_preflight_skipped":  skipPreflight,
	}
	if lease.Lease.Session.SessionID != "" {
		out["_proxy_runtime_session_hash"] = hashx.ShortSHA256(lease.Lease.Session.SessionID, 12)
	}
	if routeLabel := chainRouteLabel(chainPlan); routeLabel != "" {
		out["_proxy_runtime_chain_route"] = routeLabel
	}
	exitIP := ""
	if !skipPreflight {
		exitIP, err = s.checkProxyRuntimeExitIP(ctx, baseURL, listenerID)
		if err != nil {
			return nil, "", "", err
		}
	}
	return out, listenerID, exitIP, nil
}
