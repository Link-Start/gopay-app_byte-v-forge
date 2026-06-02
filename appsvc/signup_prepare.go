package appsvc

import (
	"context"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/envx"
	gopayapp "github.com/byte-v-forge/gopay-app/protocol/app"
)

func (s *Server) prepareSignupStartState(ctx context.Context, state stateMap, input signupStartInput) (gopayapp.DeviceFingerprint, string, error) {
	s.clearSignupState(state, "")
	s.clearLoginState(state, "")
	device, err := s.ensureDevice(state)
	if err != nil {
		return device, "", err
	}
	if envx.String("GOPAY_APP_VERSION") == "" && !strings.HasPrefix(strings.TrimSpace(device.AppVersion), "2.7.") {
		next, rawDevice, err := s.newLogonDevice()
		if err != nil {
			return device, "", err
		}
		device = next
		state["device"] = rawDevice
	}
	if probeTransactionID := stateString(state, "_signup_probe_transaction_id"); probeTransactionID != "" {
		device.TransactionID = probeTransactionID
	} else {
		state["_signup_probe_transaction_id"] = device.TransactionID
	}
	state["device"] = deviceToMap(device)
	deleteKeys(state, activeTokenKeys...)
	deleteKeys(state, activeTokenMetaKeys...)
	deleteKeys(state, tmpTokenKeys...)
	deleteKeys(state, tmpTokenMetaKeys...)
	state["_signup_phone"] = input.Phone
	state["_signup_country_code"] = input.CountryCode
	state["_signup_name"] = input.Name
	state["_signup_email"] = input.Email
	state["_signup_started_at"] = time.Now().Unix()
	state["_signup_skip_phone_probe"] = input.SkipPhoneProbe
	state["stage"] = "signup"
	delete(state, "last_error")
	if err := s.ensureProxyRuntimeSession(ctx, state, proxyRuntimeAcquireOptions{CountryCode: input.CountryCode}); err != nil {
		return device, "", err
	}
	return device, s.proxyForState(state), nil
}
