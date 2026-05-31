package app

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/byte-v-forge/common-lib/envx"
	"github.com/byte-v-forge/common-lib/randx"
)

func androidDeviceOS(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = defaultAndroidVersion
	}
	if strings.HasPrefix(strings.ToLower(value), "android") {
		parts := strings.SplitN(value, ",", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) + ", " + strings.TrimSpace(parts[1])
		}
		return value
	}
	return "Android, " + value
}

func normalizePhoneModel(value string) string {
	value = strings.TrimSpace(value)
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return value
	}
	return strings.TrimSpace(parts[0]) + ", " + strings.TrimSpace(parts[1])
}

func generatedOrStatic(static bool, staticValue string, generate func() string) string {
	if static {
		return staticValue
	}
	return generate()
}

func randomD1() string {
	raw := randomBytes(32)
	parts := make([]string, 0, len(raw))
	for _, b := range raw {
		parts = append(parts, fmt.Sprintf("%02X", b))
	}
	return strings.Join(parts, ":")
}

func randomHardwareProfile(static bool) hardwareProfile {
	if static || !envx.Bool("GOPAY_RANDOM_HARDWARE_PROFILE", false) || len(hardwareProfiles) == 0 {
		return hardwareProfile{
			AndroidVersion: defaultAndroidVersion,
			PhoneMake:      defaultPhoneMake,
			PhoneModel:     defaultPhoneModel,
			Screen:         defaultXM1Screen,
		}
	}
	return hardwareProfiles[randomIntRange(0, len(hardwareProfiles)-1)]
}

func randomAppsFlyerID() string {
	installUnixMillis := time.Now().Add(-time.Duration(randomIntRange(60, 86_400*21)) * time.Second).UnixMilli()
	return fmt.Sprintf("%d-%09d%010d", installUnixMillis, randomIntRange(100_000_000, 999_999_999), randomIntRange(0, 9_999_999_999))
}

func randomWidevineID() string {
	return base64.StdEncoding.EncodeToString(randomBytes(32))
}

func randomWiFiMAC() string {
	raw := randomBytes(6)
	raw[0] = (raw[0] | 0x02) & 0xFE
	parts := make([]string, 0, len(raw))
	for _, b := range raw {
		parts = append(parts, fmt.Sprintf("%02x", b))
	}
	return strings.Join(parts, ":")
}

func randomWiFiSSID() string {
	return defaultXM1WiFiSSID
}

func randomM1ConnectionID() string {
	return fmt.Sprintf("%d", randomIntRange(10000, 99999))
}

func randomM1SignatureTime() string {
	seenAt := time.Now().Add(-time.Duration(randomIntRange(60, 86_400*7)) * time.Second)
	return fmt.Sprint(seenAt.UnixMilli())
}

func randomFCMToken() string {
	return randomURLSafe(11) + ":APA91b" + randomURLSafe(134)
}

func randomPrivateIP() string {
	switch randomIntRange(0, 2) {
	case 0:
		return fmt.Sprintf("192.168.%d.%d", randomIntRange(0, 50), randomIntRange(2, 254))
	case 1:
		return fmt.Sprintf("10.%d.%d.%d", randomIntRange(0, 50), randomIntRange(0, 255), randomIntRange(2, 254))
	default:
		return fmt.Sprintf("172.%d.%d.%d", randomIntRange(16, 31), randomIntRange(0, 255), randomIntRange(2, 254))
	}
}

func randomURLSafe(size int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	raw := randomBytes(size)
	out := make([]byte, size)
	for idx, value := range raw {
		out[idx] = alphabet[int(value)%len(alphabet)]
	}
	return string(out)
}

func randomHex(size int) string {
	return hex.EncodeToString(randomBytes(size))
}

func randomBytes(size int) []byte {
	raw, err := randx.Bytes(size)
	if err != nil {
		raw = make([]byte, size)
		fallback := []byte(uuid.NewString())
		for i := range raw {
			raw[i] = fallback[i%len(fallback)]
		}
	}
	return raw
}

func randomIntRange(minValue int, maxValue int) int {
	if maxValue <= minValue {
		return minValue
	}
	n, err := randx.IntRange(minValue, maxValue)
	if err != nil {
		return minValue + int(time.Now().UnixNano()%int64(maxValue-minValue+1))
	}
	return n
}
