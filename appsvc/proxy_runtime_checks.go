package appsvc

import (
	"context"
	"fmt"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (s *Server) checkProxyRuntimeExitIP(ctx context.Context, baseURL string, listenerID string) (string, error) {
	var parsed proxyruntimev1.GetProxyExitIPResponse
	if err := postProxyRuntime(ctx, baseURL+"/proxy_exit_ip", &proxyruntimev1.GetProxyExitIPRequest{ListenerId: listenerID}, &parsed, 20*time.Second); err != nil {
		return "", fmt.Errorf("proxy exit ip check: %w", err)
	}
	exitIP := parsed.GetProxyExitIp()
	if exitIP.GetErrorMessage() != "" {
		return "", fmt.Errorf("proxy exit ip check: %s", exitIP.GetErrorMessage())
	}
	if exitIP.GetIp() == "" {
		return "", fmt.Errorf("proxy exit ip check returned empty ip")
	}
	return exitIP.GetIp(), nil
}

func (s *Server) checkProxyRuntimeGeo(ctx context.Context, baseURL string, ip string) (*proxyruntimev1.ProxyExitGeo, error) {
	var parsed proxyruntimev1.GetProxyExitGeoResponse
	if err := postProxyRuntime(ctx, baseURL+"/proxy_exit_geo", &proxyruntimev1.GetProxyExitGeoRequest{Ip: ip}, &parsed, 20*time.Second); err != nil {
		return nil, fmt.Errorf("proxy exit geo check: %w", err)
	}
	geo := parsed.GetProxyExitGeo()
	if geo.GetErrorMessage() != "" {
		return geo, fmt.Errorf("proxy exit geo check: %s", geo.GetErrorMessage())
	}
	return geo, nil
}

func (s *Server) checkProxyRuntimeFraud(ctx context.Context, baseURL string, ip string) (*proxyruntimev1.ProxyIPFraudCheck, error) {
	var parsed proxyruntimev1.CheckProxyIPFraudResponse
	if err := postProxyRuntime(ctx, baseURL+"/ip_fraud_check", &proxyruntimev1.CheckProxyIPFraudRequest{Ip: ip}, &parsed, 25*time.Second); err != nil {
		return nil, fmt.Errorf("proxy IP fraud check: %w", err)
	}
	check := parsed.GetCheck()
	if check.GetErrorMessage() != "" && !proxyIPFraudUnsupported(check.GetRiskLevel().String(), check.GetErrorMessage()) {
		return check, fmt.Errorf("proxy IP fraud check: %s", check.GetErrorMessage())
	}
	return check, nil
}

func (s *Server) checkGoPayProxyRuntimeConnectivity(ctx context.Context, baseURL string, listenerID string) ([]map[string]any, error) {
	results := make([]map[string]any, 0, len(goPayProxyConnectivityTargets))
	for _, target := range goPayProxyConnectivityTargets {
		var parsed proxyruntimev1.CheckProxyTargetConnectivityResponse
		if err := postProxyRuntime(ctx, baseURL+"/target_connectivity_check", &proxyruntimev1.CheckProxyTargetConnectivityRequest{ListenerId: listenerID, TargetUrl: target}, &parsed, 20*time.Second); err != nil {
			return results, fmt.Errorf("target connectivity check %s: %w", target, err)
		}
		check := parsed.GetCheck()
		result := map[string]any{
			"target_url":  target,
			"reachable":   check.GetReachable(),
			"status_code": check.GetStatusCode(),
			"latency_ms":  check.GetLatencyMs(),
		}
		if check.GetErrorMessage() != "" {
			result["error_message"] = check.GetErrorMessage()
		}
		results = append(results, result)
		if !check.GetReachable() {
			return results, fmt.Errorf("target connectivity failed: %s", target)
		}
	}
	return results, nil
}
