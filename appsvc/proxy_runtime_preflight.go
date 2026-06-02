package appsvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/hashx"
)

func (s *Server) acquireAndPreflightProxyRuntimeSession(ctx context.Context, baseURL string, state stateMap, identity string, countryCode string, forceNew bool, leaseTTL string, attempt int, requireLineProxy bool) (map[string]any, error) {
	out, listenerID, exitIP, err := s.acquireProxyRuntimeSessionWithListener(ctx, baseURL, state, identity, countryCode, forceNew, leaseTTL, attempt, false, requireLineProxy)
	if err != nil {
		return nil, err
	}
	geo, err := s.checkProxyRuntimeGeo(ctx, baseURL, exitIP)
	if err != nil {
		return nil, s.releaseFailedProxyRuntimePreflight(ctx, baseURL, out, err)
	}
	if geo.GetCountryCode() != "" && !strings.EqualFold(geo.GetCountryCode(), countryCode) {
		return nil, s.releaseFailedProxyRuntimePreflight(ctx, baseURL, out, fmt.Errorf("proxy exit country mismatch: got %s want %s", geo.GetCountryCode(), countryCode))
	}
	fraud, err := s.checkProxyRuntimeFraud(ctx, baseURL, exitIP)
	if err != nil {
		return nil, s.releaseFailedProxyRuntimePreflight(ctx, baseURL, out, err)
	}
	if proxyIPFraudRiskRejected(fraud.GetRiskLevel().String()) {
		return nil, s.releaseFailedProxyRuntimePreflight(ctx, baseURL, out, fmt.Errorf("proxy IP fraud risk rejected: level=%s score=%.0f", fraud.GetRiskLevel().String(), fraud.GetRiskScore()))
	}
	connectivity, err := s.checkGoPayProxyRuntimeConnectivity(ctx, baseURL, listenerID)
	if err != nil {
		return nil, s.releaseFailedProxyRuntimePreflight(ctx, baseURL, out, err)
	}
	out["_proxy_runtime_preflight_country_code"] = countryCode
	out["_proxy_runtime_preflight_attempt"] = attempt
	out["_proxy_runtime_exit_ip_hash"] = hashx.ShortSHA256(exitIP, 12)
	out["_proxy_runtime_exit_country_code"] = geo.GetCountryCode()
	out["_proxy_runtime_exit_region"] = geo.GetRegion()
	out["_proxy_runtime_ip_fraud_risk_level"] = normalizeProxyRiskLevel(fraud.GetRiskLevel().String())
	out["_proxy_runtime_ip_fraud_risk_score"] = fraud.GetRiskScore()
	out["_proxy_runtime_ip_network_kind"] = shortProxyEnum(fraud.GetNetworkKind().String())
	out["_proxy_runtime_ip_anonymizer_kind"] = shortProxyEnum(fraud.GetAnonymizerKind().String())
	out["_proxy_runtime_connectivity_reachable"] = true
	out["_proxy_runtime_connectivity_targets"] = connectivity
	return out, nil
}
