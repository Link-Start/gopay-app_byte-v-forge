package appsvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/hashx"
	"github.com/byte-v-forge/common-lib/stringx"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (s *Server) acquireProxyRuntimeSession(ctx context.Context, baseURL string, state stateMap, identity string, countryCode string, forceNew bool, leaseTTL string, attempt int, skipPreflight bool, requireLineProxy bool) (map[string]any, error) {
	out, _, _, err := s.acquireProxyRuntimeSessionWithListener(ctx, baseURL, state, identity, countryCode, forceNew, leaseTTL, attempt, skipPreflight, requireLineProxy)
	return out, err
}

func (s *Server) acquireProxyRuntimeSessionWithListener(ctx context.Context, baseURL string, state stateMap, identity string, countryCode string, forceNew bool, leaseTTL string, attempt int, skipPreflight bool, requireLineProxy bool) (map[string]any, string, string, error) {
	accountID := proxyRuntimeAccountID(identity)
	leaseTTL = stringx.FirstNonEmpty(leaseTTL, goPayProxyLeaseTTL)
	leaseReq := &proxyruntimev1.AcquireProxyLeaseRequest{
		AccountId: accountID,
		Purpose:   goPayProxyPurpose,
		ForceNew:  forceNew,
		Policy: &proxyruntimev1.ProxySessionPolicy{
			Mode:         proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY,
			Region:       countryCode,
			StickyTtl:    durationpb.New(proxyRuntimeLeaseDuration(leaseTTL)),
			UpstreamKind: proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP,
			RotationMode: proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_STICKY_SESSION,
			Labels:       map[string]string{"purpose": "gopay_app", "country": countryCode},
		},
		ChainPolicy: &proxyruntimev1.ProxyChainPolicy{
			CountryCode:               countryCode,
			Purpose:                   goPayProxyPurpose,
			Strategy:                  proxyruntimev1.ProxyChainStrategy_PROXY_CHAIN_STRATEGY_REGION_AWARE,
			RequireDynamicExit:        true,
			AllowDirectDynamicGateway: !requireLineProxy,
			PreferLineProxy:           true,
			MaxAttempts:               goPayProxyPreflightMaxAttempts,
		},
	}
	var lease proxyruntimev1.AcquireProxyLeaseResponse
	if err := postProxyRuntime(ctx, baseURL+"/leases/acquire", leaseReq, &lease, 30*time.Second); err != nil {
		return nil, "", "", err
	}
	egress := lease.GetEgress()
	if egress.GetHost() == "" {
		egress = lease.GetLease().GetEgress()
	}
	proxyURL, err := proxyRuntimeProxyURL(egress)
	if err != nil {
		return nil, "", "", err
	}
	listener := lease.GetLease().GetListener()
	listenerID := strings.TrimSpace(listener.GetListenerId())
	if listenerID == "" {
		return nil, "", "", fmt.Errorf("proxy-runtime returned lease without listener")
	}

	dynamicLease := lease.GetLease()
	chainPlan := firstProxyChainPlan(lease.GetChainPlan(), dynamicLease.GetChainPlan())
	out := map[string]any{
		"_gopay_proxy":                      proxyURL,
		"_gopay_account_id":                 identity,
		"_gopay_country_code":               countryCode,
		"_proxy_runtime_account_id":         accountID,
		"_proxy_runtime_lease_id":           dynamicLease.GetLeaseId(),
		"_proxy_runtime_provider_account":   dynamicLease.GetProviderAccountId(),
		"_proxy_runtime_lease_expires_at":   protoTimestampRFC3339(dynamicLease.GetExpiresAt()),
		"_proxy_runtime_listener_id":        listenerID,
		"_proxy_runtime_listener_kind":      shortProxyEnum(listener.GetKind().String()),
		"_proxy_runtime_session_started_at": time.Now().Unix(),
		"_proxy_runtime_pool_endpoints":     len(lease.GetPool().GetEndpoints()),
		"_proxy_runtime_session_rotated":    forceNew,
		"_proxy_runtime_preflight_skipped":  skipPreflight,
	}
	if sessionID := dynamicLease.GetSession().GetSessionId(); sessionID != "" {
		out["_proxy_runtime_session_hash"] = hashx.ShortSHA256(sessionID, 12)
	}
	if routeLabel := chainRouteLabel(chainPlan); routeLabel != "" {
		out["_proxy_runtime_chain_route"] = routeLabel
	}
	exitIP := ""
	if !skipPreflight {
		exitIP, err = s.checkProxyRuntimeExitIP(ctx, baseURL, listenerID)
		if err != nil {
			_ = s.releaseProxyRuntimeAccount(ctx, baseURL, accountID)
			return nil, "", "", err
		}
	}
	return out, listenerID, exitIP, nil
}

func (s *Server) releaseFailedProxyRuntimePreflight(ctx context.Context, baseURL string, data map[string]any, err error) error {
	if releaseErr := s.releaseProxyRuntimeAccount(ctx, baseURL, anyString(data["_proxy_runtime_account_id"])); releaseErr != nil {
		return fmt.Errorf("%w; release failed: %v", err, releaseErr)
	}
	return err
}
