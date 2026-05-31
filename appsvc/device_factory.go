package appsvc

import (
	"fmt"
	"time"

	"github.com/byte-v-forge/common-lib/randx"
	gopayapp "github.com/byte-v-forge/gopay-app/protocol/app"
)

func (s *Server) ensureDevice(state stateMap) (gopayapp.DeviceFingerprint, error) {
	raw := nestedMap(state["device"])
	if len(raw) > 0 {
		device := deviceFromMap(raw)
		if deviceNeedsBackfill(device) {
			next, err := gopayapp.NewDeviceFingerprint(gopayapp.DeviceConfigFromEnv())
			if err != nil {
				return gopayapp.DeviceFingerprint{}, err
			}
			device = mergeDevice(device, next)
		}
		state["device"] = deviceToMap(device)
		return device, nil
	}
	device, err := gopayapp.NewDeviceFingerprint(gopayapp.DeviceConfigFromEnv())
	if err != nil {
		return gopayapp.DeviceFingerprint{}, err
	}
	out := deviceToMap(device)
	out["profile_id"] = randomProfileID()
	out["profile_created_at"] = time.Now().Unix()
	state["device"] = out
	return device, nil
}

func ensureRandomDevice(state stateMap) (gopayapp.DeviceFingerprint, error) {
	device, err := gopayapp.NewDeviceFingerprint(gopayapp.DeviceConfig{})
	if err != nil {
		return gopayapp.DeviceFingerprint{}, err
	}
	out := deviceToMap(device)
	out["profile_id"] = randomProfileID()
	out["profile_created_at"] = time.Now().Unix()
	state["device"] = out
	return device, nil
}

func (s *Server) newLogonDevice() (gopayapp.DeviceFingerprint, map[string]any, error) {
	device, err := gopayapp.NewDeviceFingerprint(gopayapp.DeviceConfigFromEnv())
	if err != nil {
		return gopayapp.DeviceFingerprint{}, nil, err
	}
	out := deviceToMap(device)
	out["profile_id"] = randomProfileID()
	out["profile_created_at"] = time.Now().Unix()
	return device, out, nil
}

func randomProfileID() string {
	value, err := randx.Hex(8)
	if err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return value
}

func deviceNeedsBackfill(device gopayapp.DeviceFingerprint) bool {
	return device.AppID == "" ||
		device.UniqueID == "" ||
		device.TLSProfileName == "" ||
		device.M1Hardware == "" ||
		device.IMEI == "" ||
		device.IPAddress == "" ||
		device.FirebaseID == "" ||
		device.AdvertisingID == "" ||
		device.AppSetID == "" ||
		device.M1SignatureTime == ""
}
