package app

import (
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/envx"
	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/common-lib/stringx"
)

const (
	AuthBaseURL     = "https://accounts.goto-products.com"
	CustomerBaseURL = "https://customer.gopayapi.com"
	GojekBaseURL    = "https://api.gojekapi.com"
)

type Config struct {
	Token                 string
	ProxyURL              string
	Timeout               time.Duration
	HTTPClient            *http.Client
	Device                DeviceFingerprint
	DeviceConfig          DeviceConfig
	SignVersion           string
	LegacyHMACKey         string
	DisplayEncoderKey     string
	DisplayEncoderID      string
	SignedMsgTemplatePath string
	Logger                httpjson.Logger
	DebugHTTP             bool
}

func ConfigFromEnv(token string) Config {
	return Config{
		Token:        token,
		ProxyURL:     envx.String("GOPAY_PROXY_URL"),
		DeviceConfig: DeviceConfigFromEnv(),
		DebugHTTP:    envx.Bool("GOPAY_APP_DEBUG_HTTP_REQUESTS", false),
		SignVersion:  envx.StringDefault("GOPAY_SIGN_VERSION", defaultGoPaySignVersion),
		LegacyHMACKey: stringx.FirstNonEmpty(
			envx.String("GOPAY_LEGACY_DISPLAY_ENCODER_KEY"),
			envx.String("GOPAY_HMAC_KEY"),
			defaultGoPayLegacyDisplayEncoderKey,
		),
		DisplayEncoderKey:     envx.StringDefault("GOPAY_DISPLAY_ENCODER_KEY", defaultGoPayDisplayEncoderKey),
		DisplayEncoderID:      envx.StringDefault("GOPAY_DISPLAY_ENCODER_ID", defaultGoPayDisplayEncoderID),
		SignedMsgTemplatePath: envx.String("GOPAY_SIGNED_MSG_TEMPLATE"),
	}
}
