package appsvc

import (
	"context"
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/jsonx"
)

func (s *Server) refreshPINSetupFromProfile(ctx context.Context, client anyClient, state stateMap) (bool, bool, string) {
	resp, err := client.Get(ctx, "/v1/users/profile")
	if err != nil {
		return false, false, err.Error()
	}
	if resp.StatusCode != http.StatusOK {
		return false, false, apiError("pin setup check failed", resp)
	}
	pinSetup, ok := pinSetupFlagFromProfileData(resp.Data())
	if !ok {
		return false, false, "is_pin_setup missing"
	}
	updatePINSetupState(state, pinSetup)
	return pinSetup, true, ""
}

func updatePINSetupState(state stateMap, pinSetup bool) {
	now := time.Now().Unix()
	state["pin_setup"] = pinSetup
	state["pin_setup_checked_at"] = now
	if pinSetup {
		state["pin_setup_at"] = now
		return
	}
	delete(state, "pin_setup_at")
}

func pinSetupFlagFromProfileData(value any) (bool, bool) {
	wanted := map[string]struct{}{
		jsonx.NormalizeKey("is_pin_setup"): {},
		jsonx.NormalizeKey("isPinSetup"):   {},
	}
	var walk func(any) (bool, bool)
	walk = func(current any) (bool, bool) {
		if obj, ok := jsonx.Object(current); ok {
			for key, item := range obj {
				if _, ok := wanted[jsonx.NormalizeKey(key)]; ok {
					return anyBool(item), true
				}
			}
			for _, item := range obj {
				if value, ok := walk(item); ok {
					return value, true
				}
			}
			return false, false
		}
		if items, ok := current.([]any); ok {
			for _, item := range items {
				if value, ok := walk(item); ok {
					return value, true
				}
			}
		}
		return false, false
	}
	return walk(value)
}
