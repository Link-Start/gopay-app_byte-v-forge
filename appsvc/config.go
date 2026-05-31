package appsvc

import (
	"github.com/byte-v-forge/common-lib/envx"
	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/paymentsvc"

	"os"
	"strings"
	"time"
)

type Config struct {
	Port                       string
	HTTPListenAddr             string
	DashboardStaticDir         string
	N8NWebhookBaseURL          string
	StateRedisURL              string
	StateKeyPrefix             string
	StateTTL                   time.Duration
	SignupAuthUUID             string
	PINClientID                string
	GotoClientID               string
	GotoClientSecret           string
	ProxyRuntimeHTTPAddr       string
	SignupInitiateJitterMin    time.Duration
	SignupInitiateJitterMax    time.Duration
	SignupRateLimitCooldown    time.Duration
	OTPTimeout                 time.Duration
	TokenRefreshMinTTL         time.Duration
	ChangePhoneConfirmTimeout  time.Duration
	ChangePhoneConfirmInterval time.Duration
	EnvelopeShortlinkTimeout   time.Duration
	ChangePhoneCountrySync     bool
	MinBalanceRp               int64
	OTPWebhookListenAddr       string
	OTPSubmitURL               string
	Payment                    paymentsvc.Config
}

func ConfigFromEnv() Config {
	return Config{
		Port:                       stringx.FirstNonEmpty(os.Getenv("GOPAY_APP_PORT"), "50051"),
		HTTPListenAddr:             stringx.FirstNonEmpty(os.Getenv("GOPAY_HTTP_LISTEN_ADDR"), ":8080"),
		DashboardStaticDir:         stringx.FirstNonEmpty(os.Getenv("GOPAY_DASHBOARD_STATIC_DIR"), "/app/dashboard/gopay"),
		N8NWebhookBaseURL:          strings.TrimRight(strings.TrimSpace(os.Getenv("GOPAY_N8N_WEBHOOK_BASE_URL")), "/"),
		StateRedisURL:              strings.TrimSpace(os.Getenv("GOPAY_STATE_REDIS_URL")),
		StateKeyPrefix:             stringx.FirstNonEmpty(os.Getenv("GOPAY_STATE_KEY_PREFIX"), "byte-v-forge:gopay-app:state"),
		StateTTL:                   envx.PositiveDurationSeconds("GOPAY_STATE_TTL_SECONDS", 7*24*time.Hour),
		SignupAuthUUID:             "bb648413-b637-443a-8ebf-176cf9b5dc32",
		PINClientID:                "6d11d261d7ae462dbd4be0dc5f36a697-MFAGOJEK",
		GotoClientID:               "gopay:consumer:app",
		GotoClientSecret:           strings.TrimSpace(os.Getenv("GOTO_SSO_CLIENT_SECRET")),
		OTPTimeout:                 180 * time.Second,
		TokenRefreshMinTTL:         900 * time.Second,
		ChangePhoneConfirmTimeout:  8 * time.Second,
		ChangePhoneConfirmInterval: time.Second,
		EnvelopeShortlinkTimeout:   10 * time.Second,
		ProxyRuntimeHTTPAddr:       strings.TrimSpace(os.Getenv("PROXY_RUNTIME_HTTP_ADDR")),
		SignupInitiateJitterMin:    envx.NonNegativeDurationSeconds("GOPAY_SIGNUP_INITIATE_JITTER_MIN_SECONDS", 8*time.Second),
		SignupInitiateJitterMax:    envx.NonNegativeDurationSeconds("GOPAY_SIGNUP_INITIATE_JITTER_MAX_SECONDS", 25*time.Second),
		SignupRateLimitCooldown:    envx.NonNegativeDurationSeconds("GOPAY_SIGNUP_RATE_LIMIT_COOLDOWN_SECONDS", 900*time.Second),
		MinBalanceRp:               1,
		OTPWebhookListenAddr:       stringx.FirstNonEmpty(os.Getenv("GOPAY_OTP_WEBHOOK_LISTEN_ADDR"), ":8081"),
		OTPSubmitURL:               stringx.FirstNonEmpty(os.Getenv("GOPAY_OTP_SUBMIT_URL"), "http://localhost:8080/api/gopay/otp/submit"),
		Payment:                    paymentsvc.ConfigFromEnv(),
	}
}
