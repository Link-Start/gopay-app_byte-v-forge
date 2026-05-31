package appsvc

import (
	"context"
	"fmt"

	"github.com/byte-v-forge/common-lib/stringx"
	gopayapp "github.com/byte-v-forge/gopay-app/protocol/app"
)

func (s *Server) newClientWithState(ctx context.Context, state stateMap, requireToken bool) (*gopayapp.ClientSet, error) {
	if requireToken {
		refresh := s.ensureAccessToken(ctx, state, s.cfg.TokenRefreshMinTTL, false)
		if !anyBool(refresh["success"]) && !tokenUsable(state, "token", 0) {
			return nil, fmt.Errorf("%s", stringx.FirstNonEmpty(anyString(refresh["error"]), "token refresh failed"))
		}
	}
	if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{}); err != nil {
		return nil, err
	}
	device, err := s.ensureDevice(state)
	if err != nil {
		return nil, err
	}
	return s.newClient(ctx, stateString(state, "token"), s.proxyForState(state), device)
}
