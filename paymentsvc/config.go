package paymentsvc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/envx"
	"github.com/byte-v-forge/common-lib/stringx"
)

const (
	defaultMidtransClientID = "Mid-client-3TX8nUa-f_RgNrky"
	defaultTokenization     = "true"
	defaultBrowserLocale    = "en-US"
	defaultPINLocale        = "id"
	defaultBrowserPlatform  = "Windows"
	defaultTLSProfile       = "chrome_146"
)

type Config struct {
	PaymentProfile   requestProfile
	MidtransClientID string
}

type requestProfile struct {
	Name           string `json:"name"`
	ProxyURL       string `json:"proxy_url"`
	TLSProfile     string `json:"tls_profile"`
	UserAgent      string `json:"user_agent"`
	SecCHUA        string `json:"sec_ch_ua"`
	SecCHPlatform  string `json:"sec_ch_ua_platform"`
	AcceptLanguage string `json:"accept_language"`
	OAILanguage    string `json:"oai_language"`
	Locale         string `json:"locale"`
	DeviceID       string `json:"device_id"`
	Platform       string `json:"platform"`
	PINLocale      string `json:"pin_locale"`
}

func ConfigFromEnv() Config {
	return Config{
		PaymentProfile:   requestProfileFromEnv("GOPAY_PAYMENT_PROFILE_JSON", defaultRequestProfile("payment")),
		MidtransClientID: envx.StringDefault("GOPAY_MIDTRANS_CLIENT_ID", defaultMidtransClientID),
	}
}

func defaultRequestProfile(name string) requestProfile {
	return requestProfile{Name: name, TLSProfile: defaultTLSProfile, Locale: defaultBrowserLocale, Platform: defaultBrowserPlatform, PINLocale: defaultPINLocale}
}

func requestProfileFromEnv(envName string, fallback requestProfile) requestProfile {
	profile := fallback
	if raw := envx.String(envName); raw != "" {
		if err := json.Unmarshal([]byte(raw), &profile); err != nil {
			panic(fmt.Sprintf("invalid %s: %v", envName, err))
		}
	}
	return profile.withDefaults(fallback)
}

func (p requestProfile) withDefaults(fallback requestProfile) requestProfile {
	p.Name = stringx.FirstNonEmpty(p.Name, fallback.Name)
	p.ProxyURL = stringx.FirstNonEmpty(p.ProxyURL, fallback.ProxyURL)
	p.TLSProfile = stringx.FirstNonEmpty(p.TLSProfile, fallback.TLSProfile, defaultTLSProfile)
	p.Locale = stringx.FirstNonEmpty(p.Locale, fallback.Locale, defaultBrowserLocale)
	p.Platform = stringx.FirstNonEmpty(p.Platform, fallback.Platform, defaultBrowserPlatform)
	p.PINLocale = stringx.FirstNonEmpty(p.PINLocale, fallback.PINLocale, defaultPINLocale)
	p.UserAgent = strings.TrimSpace(p.UserAgent)
	p.SecCHUA = strings.TrimSpace(p.SecCHUA)
	p.SecCHPlatform = strings.TrimSpace(p.SecCHPlatform)
	p.AcceptLanguage = strings.TrimSpace(p.AcceptLanguage)
	p.OAILanguage = strings.TrimSpace(p.OAILanguage)
	p.DeviceID = strings.TrimSpace(p.DeviceID)
	return p
}

func (p requestProfile) fingerprint() browserFingerprint { return browserFingerprintFromProfile(p) }
