package appsvc

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/hashx"
	"github.com/byte-v-forge/common-lib/stringx"
)

func proxyRuntimeAPIBase(value string) string {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return ""
	}
	if strings.HasSuffix(value, "/api/proxy-runtime") || strings.HasSuffix(value, "/proxy") {
		return value
	}
	return value + "/api/proxy-runtime"
}

func proxyRuntimeAccountID(identity string) string {
	identity = strings.TrimSpace(identity)
	if identity == "" {
		return ""
	}
	if strings.HasPrefix(identity, "gopay-app-") {
		return identity
	}
	return "gopay-app-" + hashx.ShortSHA256(identity, 20)
}

func (s *Server) bindGoPayAccountIdentity(state stateMap, explicit string) string {
	identity := strings.TrimSpace(explicit)
	if identity == "" {
		identity = stateString(state, "_gopay_account_id")
	}
	if identity == "" {
		identity = "local"
	}
	if state != nil {
		state["_gopay_account_id"] = identity
		state["_proxy_runtime_account_id"] = proxyRuntimeAccountID(identity)
	}
	return identity
}

func normalizeGoPayProxyCountryCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "+")
	switch value {
	case "", "62", "ID", "IDN", "INDONESIA":
		return goPayProxyCountryCode
	default:
		return value
	}
}

func proxyRuntimeProxyURL(endpoint proxyRuntimeEndpoint) (string, error) {
	if endpoint.Host == "" || endpoint.Port <= 0 {
		return "", fmt.Errorf("proxy-runtime returned invalid lease egress")
	}
	scheme := "http"
	switch endpoint.Protocol {
	case "PROXY_PROTOCOL_SOCKS5", "SOCKS5", "socks5":
		scheme = "socks5"
	}
	return (&url.URL{Scheme: scheme, Host: fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)}).String(), nil
}

func proxyRuntimeLeaseActive(state stateMap) bool {
	if stateString(state, "_proxy_runtime_lease_id") == "" {
		return false
	}
	expiresAt := stateString(state, "_proxy_runtime_lease_expires_at")
	if expiresAt == "" {
		return false
	}
	parsed, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return false
	}
	return time.Now().Add(30 * time.Second).Before(parsed)
}

func proxyIPFraudRiskRejected(level string) bool {
	switch normalizeProxyRiskLevel(level) {
	case "HIGH", "CRITICAL":
		return true
	default:
		return false
	}
}

func proxyIPFraudUnsupported(level string, message string) bool {
	if normalizeProxyRiskLevel(level) == "UNSUPPORTED" {
		return true
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(message)), "unsupported")
}

func normalizeProxyRiskLevel(level string) string {
	level = strings.ToUpper(strings.TrimSpace(level))
	level = strings.TrimPrefix(level, "PROXY_IP_FRAUD_RISK_LEVEL_")
	return level
}

func shortProxyEnum(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	for _, prefix := range []string{"PROXY_IP_NETWORK_KIND_", "PROXY_IP_ANONYMIZER_KIND_", "PROXY_PROTOCOL_", "PROXY_UPSTREAM_KIND_", "PROXY_ROTATION_MODE_", "EGRESS_LISTENER_KIND_"} {
		value = strings.TrimPrefix(value, prefix)
	}
	return value
}

func firstMap(values ...map[string]any) map[string]any {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func chainRouteLabel(chainPlan map[string]any) string {
	if len(chainPlan) == 0 {
		return ""
	}
	line := nestedMap(chainPlan["line"])
	dynamicGateway := nestedMap(chainPlan["dynamic_gateway"])
	lineName := stringx.FirstNonEmpty(anyString(line["display_name"]), anyString(line["node_id"]), anyString(line["source_id"]))
	gatewayName := stringx.FirstNonEmpty(anyString(dynamicGateway["display_name"]), anyString(dynamicGateway["gateway_id"]), anyString(dynamicGateway["provider_id"]))
	switch {
	case lineName != "" && gatewayName != "":
		return lineName + " -> " + gatewayName
	case gatewayName != "":
		return gatewayName
	default:
		return lineName
	}
}
