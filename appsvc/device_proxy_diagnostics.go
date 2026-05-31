package appsvc

import (
	"strings"

	"github.com/byte-v-forge/common-lib/hashx"
)

func (s *Server) deviceProxyDiagnostics(state stateMap) map[string]any {
	data := map[string]any{}
	proxyURL := stateString(state, "_gopay_proxy")
	if proxyURL != "" {
		data["proxy_hash"] = hashx.ShortSHA256(proxyURL, 12)
	}
	if hash := stateString(state, "_proxy_runtime_session_hash"); hash != "" {
		data["proxy_runtime_session_hash"] = hash
		data["proxy_runtime_pool_endpoints"] = anyInt(state["_proxy_runtime_pool_endpoints"])
	}
	if rotated, ok := state["_proxy_runtime_session_rotated"].(bool); ok {
		data["proxy_runtime_session_rotated"] = rotated
	}
	for _, key := range []string{
		"_gopay_account_id",
		"_gopay_country_code",
		"_proxy_runtime_account_id",
		"_proxy_runtime_lease_id",
		"_proxy_runtime_lease_expires_at",
		"_proxy_runtime_listener_id",
		"_proxy_runtime_chain_route",
		"_proxy_runtime_exit_ip_hash",
		"_proxy_runtime_exit_country_code",
		"_proxy_runtime_exit_region",
		"_proxy_runtime_ip_fraud_risk_level",
		"_proxy_runtime_ip_fraud_risk_score",
		"_proxy_runtime_ip_network_kind",
		"_proxy_runtime_ip_anonymizer_kind",
		"_proxy_runtime_connectivity_reachable",
		"_proxy_runtime_connectivity_targets",
		"_proxy_runtime_preflight_skipped",
		"_proxy_runtime_preflight_attempts",
		"_gopay_profile_ephemeral",
	} {
		if value, ok := state[key]; ok {
			data[strings.TrimPrefix(key, "_")] = value
		}
	}
	if fp := deviceFingerprintForState(state); fp != "" {
		data["device_fingerprint"] = fp
	}
	return data
}

func deviceFingerprintForState(state stateMap) string {
	device := nestedMap(state["device"])
	if len(device) == 0 {
		return ""
	}
	out := []string{}
	addPlain := func(label, key string) {
		if value := anyString(device[key]); value != "" {
			out = append(out, label+"="+value)
		}
	}
	addHash := func(label, key string) {
		if value := anyString(device[key]); value != "" {
			out = append(out, label+"#"+hashx.ShortSHA256(value, 12))
		}
	}
	addPlain("profile", "profile_id")
	addPlain("make", "x-phonemake")
	addPlain("model", "x-phonemodel")
	addPlain("os", "x-deviceos")
	addPlain("screen", "m1_screen")
	addPlain("tls", "tls_profile")
	addHash("uid", "x-uniqueid")
	addHash("session", "x-session-id")
	addHash("tx", "transaction-id")
	addHash("d1", "d1")
	addHash("conn", "m1_connection_id")
	addHash("widevine", "m1_widevine_id")
	addHash("wifi", "m1_wifi_mac")
	addHash("ssid", "m1_wifi_ssid")
	addHash("sig", "m1_signature")
	addHash("sig_time", "m1_signature_time")
	addHash("firebase", "m1_firebase_app_instance_id")
	addHash("uuid", "m1_device_uuid")
	addHash("adid", "advertising_id")
	addHash("appset", "app_set_id")
	addHash("devtoken", "x-devicetoken")
	addHash("imei", "x-imei")
	addHash("ip", "x-ipaddress")
	if parsed := deviceFromMap(device); parsed.AppID != "" {
		out = append(out, "x_m1#"+hashx.ShortSHA256(parsed.XM1(), 12))
	}
	return strings.Join(out, "/")
}
