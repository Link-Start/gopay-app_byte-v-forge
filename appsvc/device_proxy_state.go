package appsvc

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/hashx"
)

func (s *Server) generateDeviceProxyState(ctx context.Context, accountID string, countryCode string, forceNew bool, skipPreflight bool, ephemeralProfile bool) (stateMap, error) {
	return s.generateDeviceProxyStateWithLeaseTTL(ctx, accountID, countryCode, forceNew, skipPreflight, ephemeralProfile, "")
}

func (s *Server) generateDeviceProxyStateWithLeaseTTL(ctx context.Context, accountID string, countryCode string, forceNew bool, skipPreflight bool, ephemeralProfile bool, leaseTTLOverride string) (stateMap, error) {
	identity := strings.TrimSpace(accountID)
	if identity == "" {
		identity = "local"
	}
	state := stateMap{}
	if ephemeralProfile {
		identity = "ephemeral:" + randomProfileID()
		state["_gopay_profile_ephemeral"] = true
	} else {
		state = s.loadAccountProfile(ctx, identity)
	}
	s.bindGoPayAccountIdentity(state, identity)
	state["_gopay_country_code"] = normalizeGoPayProxyCountryCode(countryCode)
	var err error
	if ephemeralProfile {
		_, err = ensureRandomDevice(state)
	} else {
		_, err = s.ensureDevice(state)
	}
	if err != nil {
		return state, err
	}
	leaseTTL := goPayProxyLeaseTTL
	if ephemeralProfile {
		leaseTTL = goPayProxyProbeLeaseTTL
	}
	if value := strings.TrimSpace(leaseTTLOverride); value != "" {
		leaseTTL = value
	}
	if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{AccountID: identity, CountryCode: countryCode, ForceNew: forceNew, SkipPreflight: skipPreflight, LeaseTTL: leaseTTL}); err != nil {
		return state, err
	}
	if !ephemeralProfile {
		_ = s.saveAccountProfile(ctx, identity, state)
	}
	return state, nil
}

func (s *Server) loadAccountProfile(ctx context.Context, identity string) stateMap {
	key := accountProfileStateKey(identity)
	if key == "" || s.store == nil {
		return stateMap{}
	}
	raw, err := s.store.Load(ctx, key)
	if err != nil {
		return stateMap{}
	}
	state, err := parseState(raw)
	if err != nil {
		return stateMap{}
	}
	return state
}

func (s *Server) saveAccountProfile(ctx context.Context, identity string, state stateMap) error {
	key := accountProfileStateKey(identity)
	if key == "" || s.store == nil {
		return nil
	}
	profile := stateMap{}
	if device := nestedMap(state["device"]); len(device) > 0 {
		profile["device"] = device
	}
	if accountID := stateString(state, "_gopay_account_id"); accountID != "" {
		profile["_gopay_account_id"] = accountID
	}
	if countryCode := stateString(state, "_gopay_country_code"); countryCode != "" {
		profile["_gopay_country_code"] = countryCode
	}
	if proxyAccountID := stateString(state, "_proxy_runtime_account_id"); proxyAccountID != "" {
		profile["_proxy_runtime_account_id"] = proxyAccountID
	}
	if len(profile) == 0 {
		return nil
	}
	_, err := s.store.Save(ctx, key, stateJSON(profile))
	return err
}

func accountProfileStateKey(identity string) string {
	identity = strings.TrimSpace(identity)
	if identity == "" {
		return ""
	}
	return "profile:" + hashx.ShortSHA256(identity, 24)
}
