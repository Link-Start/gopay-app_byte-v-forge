package appsvc

import (
	"context"
	"fmt"
	"time"
)

func (s *Server) checkProxyRuntimeExitIP(ctx context.Context, baseURL string, listenerID string) (string, error) {
	var parsed proxyRuntimeExitIPResponse
	if err := postProxyRuntime(ctx, baseURL+"/proxy_exit_ip", map[string]any{"listener_id": listenerID}, &parsed, 20*time.Second); err != nil {
		return "", fmt.Errorf("proxy exit ip check: %w", err)
	}
	if parsed.ProxyExitIP.ErrorMessage != "" {
		return "", fmt.Errorf("proxy exit ip check: %s", parsed.ProxyExitIP.ErrorMessage)
	}
	if parsed.ProxyExitIP.IP == "" {
		return "", fmt.Errorf("proxy exit ip check returned empty ip")
	}
	return parsed.ProxyExitIP.IP, nil
}

func (s *Server) checkProxyRuntimeGeo(ctx context.Context, baseURL string, ip string) (proxyRuntimeGeoResponse, error) {
	var parsed proxyRuntimeGeoResponse
	if err := postProxyRuntime(ctx, baseURL+"/proxy_exit_geo", map[string]any{"ip": ip}, &parsed, 20*time.Second); err != nil {
		return parsed, fmt.Errorf("proxy exit geo check: %w", err)
	}
	if parsed.ProxyExitGeo.ErrorMessage != "" {
		return parsed, fmt.Errorf("proxy exit geo check: %s", parsed.ProxyExitGeo.ErrorMessage)
	}
	return parsed, nil
}

func (s *Server) checkProxyRuntimeFraud(ctx context.Context, baseURL string, ip string) (proxyRuntimeFraudResponse, error) {
	var parsed proxyRuntimeFraudResponse
	if err := postProxyRuntime(ctx, baseURL+"/ip_fraud_check", map[string]any{"ip": ip}, &parsed, 25*time.Second); err != nil {
		return parsed, fmt.Errorf("proxy IP fraud check: %w", err)
	}
	if parsed.Check.ErrorMessage != "" && !proxyIPFraudUnsupported(parsed.Check.RiskLevel, parsed.Check.ErrorMessage) {
		return parsed, fmt.Errorf("proxy IP fraud check: %s", parsed.Check.ErrorMessage)
	}
	return parsed, nil
}

func (s *Server) checkGoPayProxyRuntimeConnectivity(ctx context.Context, baseURL string, listenerID string) ([]map[string]any, error) {
	results := make([]map[string]any, 0, len(goPayProxyConnectivityTargets))
	for _, target := range goPayProxyConnectivityTargets {
		var parsed proxyRuntimeConnectivityResponse
		if err := postProxyRuntime(ctx, baseURL+"/target_connectivity_check", map[string]any{"listener_id": listenerID, "target_url": target}, &parsed, 20*time.Second); err != nil {
			return results, fmt.Errorf("target connectivity check %s: %w", target, err)
		}
		result := map[string]any{
			"target_url":  target,
			"reachable":   parsed.Check.Reachable,
			"status_code": parsed.Check.StatusCode,
			"latency_ms":  parsed.Check.LatencyMS,
		}
		if parsed.Check.ErrorMessage != "" {
			result["error_message"] = parsed.Check.ErrorMessage
		}
		results = append(results, result)
		if !parsed.Check.Reachable {
			return results, fmt.Errorf("target connectivity failed: %s", target)
		}
	}
	return results, nil
}
