package app

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func setRequestHost(headers http.Header, rawURL string) {
	parsed, err := url.Parse(rawURL)
	if err == nil && parsed.Host != "" {
		setHeader(headers, "Host", parsed.Host)
	}
}

func mergeHeaderValues(headers http.Header, extra http.Header) {
	for key, values := range extra {
		setHeaderValues(headers, key, values)
	}
}

func defaultSignedHeaders(device DeviceFingerprint, xM1 string, hasBody bool) http.Header {
	headers := http.Header{}
	setBaseHeaders(headers, device, xM1, hasBody)
	setHeader(headers, "User-uuid", device.UserUUID)
	setHeader(headers, "X-DeviceToken", device.DeviceToken)
	setHeader(headers, "X-IMEI", device.IMEI)
	setHeader(headers, "X-IpAddress", device.IPAddress)
	setHeader(headers, "X-Location", device.Location)
	setHeader(headers, "X-Location-Accuracy", device.LocationAccuracy)
	setHeader(headers, "Gojek-Country-Code", device.GojekCountryCode)
	setHeader(headers, "X-Dark-Mode", "false")
	return headers
}

func setHeader(headers http.Header, key, value string) {
	deleteHeader(headers, key)
	headers[key] = []string{value}
}

func setHeaderValues(headers http.Header, key string, values []string) {
	deleteHeader(headers, key)
	headers[key] = append([]string(nil), values...)
}

func deleteHeader(headers http.Header, key string) {
	for existing := range headers {
		if strings.EqualFold(existing, key) {
			delete(headers, existing)
		}
	}
}

func setBaseHeaders(headers http.Header, device DeviceFingerprint, xM1 string, hasBody bool) {
	setHeader(headers, "X-AppVersion", device.AppVersion)
	setHeader(headers, "X-AppId", device.AppID)
	setHeader(headers, "X-AppType", device.AppType)
	setHeader(headers, "Accept", "application/json")
	setHeader(headers, "User-Agent", device.UserAgent)
	setHeader(headers, "D1", device.D1)
	setHeader(headers, "X-Session-ID", device.SessionID)
	setHeader(headers, "X-Platform", device.Platform)
	setHeader(headers, "X-UniqueId", device.UniqueID)
	setHeader(headers, "X-User-Type", device.UserType)
	setHeader(headers, "X-DeviceOS", device.DeviceOS)
	setHeader(headers, "X-PhoneMake", device.PhoneMake)
	setHeader(headers, "X-PushTokenType", device.PushTokenType())
	setHeader(headers, "X-DeviceToken", device.DeviceToken)
	setHeader(headers, "X-IMEI", device.IMEI)
	setHeader(headers, "X-IpAddress", device.IPAddress)
	setHeader(headers, "X-PhoneModel", device.PhoneModel)
	setHeader(headers, "Accept-Language", defaultAcceptLanguage)
	setHeader(headers, "X-User-Locale", defaultUserLocale)
	setHeader(headers, "X-M1", xM1)
	setHeader(headers, "X-E2", device.XE2)
	setHeader(headers, "AdjTs", device.AdjTS)
	if hasBody {
		setHeader(headers, "Content-Type", "application/json")
	}
}

func appHeaders(device DeviceFingerprint, xM1 string, hasBody bool) http.Header {
	headers := http.Header{}
	setHeader(headers, "Accept-Encoding", "gzip")
	setHeader(headers, "Gojek-Service-Area", "1")
	setHeader(headers, "Country-Code", device.GojekCountryCode)
	setHeader(headers, "X-AppVersion", device.AppVersion)
	setHeader(headers, "X-M1", xM1)
	setHeader(headers, "Gojek-Country-Code", device.GojekCountryCode)
	setHeader(headers, "X-Request-ID", newTimeUUIDString())
	setHeader(headers, "X-UniqueId", device.UniqueID)
	setHeader(headers, "X-IMEI", device.IMEI)
	setHeader(headers, "X-IpAddress", device.IPAddress)
	setHeader(headers, "X-PhoneMake", device.PhoneMake)
	setHeader(headers, "X-Help-Version", device.AppVersion)
	setHeader(headers, "X-DeviceToken", device.DeviceToken)
	setHeader(headers, "X-Location", device.Location)
	setHeader(headers, "X-Location-Accuracy", device.LocationAccuracy)
	setHeader(headers, "X-DeviceOS", device.DeviceOS)
	setHeader(headers, "X-User-Type", device.UserType)
	setHeader(headers, "User-Agent", device.UserAgent)
	setHeader(headers, "X-AppId", device.AppID)
	setHeader(headers, "Gojek-Timezone", defaultTimezone)
	setHeader(headers, "X-AuthSDK-Version", defaultAuthSDKVersion)
	setHeader(headers, "X-AppType", device.AppType)
	setHeader(headers, "X-User-Locale", defaultUserLocale)
	setHeader(headers, "X-E2", device.XE2)
	setHeader(headers, "X-CVSDK-Version", defaultCVSDKVersion)
	setHeader(headers, "Accept-Language", defaultAcceptLanguage)
	setHeader(headers, "Transaction-ID", device.TransactionID)
	setHeader(headers, "X-PhoneModel", device.PhoneModel)
	setHeader(headers, "X-Platform", device.Platform)
	if hasBody {
		setHeader(headers, "Content-Type", "application/json")
	}
	return headers
}

func gotoAuthHeaders(device DeviceFingerprint, xM1 string, hasBody bool) http.Header {
	headers := http.Header{}
	setHeader(headers, "Accept-Encoding", "gzip")
	setHeader(headers, "X-CVSDK-Version", defaultCVSDKVersion)
	setHeader(headers, "Gojek-Service-Area", "1")
	setHeader(headers, "X-Request-ID", newTimeUUIDString())
	setHeader(headers, "Country-Code", device.GojekCountryCode)
	setHeader(headers, "X-AppVersion", device.AppVersion)
	setHeader(headers, "X-M1", xM1)
	setHeader(headers, "Gojek-Country-Code", device.GojekCountryCode)
	setHeader(headers, "X-UniqueId", device.UniqueID)
	setHeader(headers, "X-PhoneMake", device.PhoneMake)
	setHeader(headers, "X-Help-Version", device.AppVersion)
	setHeader(headers, "User-Agent", device.UserAgent)
	setHeader(headers, "X-DeviceOS", device.DeviceOS)
	setHeader(headers, "X-User-Type", device.UserType)
	setHeader(headers, "X-AppId", device.AppID)
	setHeader(headers, "Gojek-Timezone", defaultTimezone)
	setHeader(headers, "X-AuthSDK-Version", defaultAuthSDKVersion)
	setHeader(headers, "X-AppType", device.AppType)
	setHeader(headers, "X-User-Locale", defaultUserLocale)
	setHeader(headers, "X-DeviceToken", device.DeviceToken)
	setHeader(headers, "X-E2", device.XE2)
	setHeader(headers, "Accept-Language", defaultAcceptLanguage)
	setHeader(headers, "Transaction-ID", device.TransactionID)
	setHeader(headers, "X-PhoneModel", device.PhoneModel)
	setHeader(headers, "X-Platform", device.Platform)
	if hasBody {
		setHeader(headers, "Content-Type", "application/json")
	}
	return headers
}

func supportCustomerHeaders(device DeviceFingerprint, xM1 string, hasBody bool) http.Header {
	headers := http.Header{}
	setHeader(headers, "Accept-Encoding", "gzip")
	setHeader(headers, "Gojek-Service-Area", "1")
	setHeader(headers, "Country-Code", device.GojekCountryCode)
	setHeader(headers, "Support-Request-Id", newTimeUUIDString())
	setHeader(headers, "X-AppVersion", device.AppVersion)
	setHeader(headers, "X-M1", xM1)
	setHeader(headers, "Gojek-Country-Code", device.GojekCountryCode)
	setHeader(headers, "X-UniqueId", device.UniqueID)
	setHeader(headers, "X-PhoneMake", device.PhoneMake)
	setHeader(headers, "X-Help-Version", device.AppVersion)
	setHeader(headers, "User-Agent", device.UserAgent)
	setHeader(headers, "X-DeviceOS", device.DeviceOS)
	setHeader(headers, "X-User-Type", device.UserType)
	setHeader(headers, "X-AppId", device.AppID)
	setHeader(headers, "Gojek-Timezone", defaultTimezone)
	setHeader(headers, "X-AppType", device.AppType)
	setHeader(headers, "X-User-Locale", defaultUserLocale)
	setHeader(headers, "X-DeviceToken", device.DeviceToken)
	setHeader(headers, "X-E2", device.XE2)
	setHeader(headers, "Accept-Language", defaultAcceptLanguage)
	setHeader(headers, "X-PhoneModel", device.PhoneModel)
	setHeader(headers, "Support-SDK-Version", defaultSupportSDK)
	setHeader(headers, "X-Platform", device.Platform)
	if hasBody {
		setHeader(headers, "Content-Type", "application/json")
	}
	return headers
}

func newTimeUUIDString() string {
	value, err := uuid.NewUUID()
	if err == nil {
		return value.String()
	}
	return uuid.NewString()
}

var gopayCustomerSlimGetPaths = map[string]bool{
	"/v1/users/profile":            true,
	"/v1/payment-options/balances": true,
	"/v1/payment-options/profiles": true,
	"/v1/user/wallet-card/balance": true,
}

var gopayCustomerAppHeaderPaths = map[string]bool{
	"/v1/users/profile":                               true,
	"/v1/qris/payments":                               true,
	"/v2/customer/payment-options/checkout/list":      true,
	"/v1/customer/payment-options/settings/last-used": true,
	"/v1/promotions/evaluate":                         true,
	"/api/v1/festival-envelopes/claim":                true,
	"/api/v1/users/deactivate":                        true,
	"/api/v1/users/deactivate/check":                  true,
	"/api/v1/users/pin/challenges":                    true,
	"/api/v1/users/pin/tokens":                        true,
	"/api/v1/users/pin/tokens/nb":                     true,
	"/api/v1/users/pins/allowed":                      true,
	"/api/v2/users/pins/setup/tokens":                 true,
	"/cvs/v1/methods":                                 true,
	"/cvs/v1/initiate":                                true,
	"/cvs/v1/verify":                                  true,
}

var gojekActivityPaths = map[string]bool{
	"/v5/customers": true,
	"/v2/otp/retry": true,
	"/v5/customers/verificationUpdateProfile": true,
	"/gojek/v2/customer":                      true,
}

var gojekAppHeaderPaths = map[string]bool{
	"/courier/v1/token":    true,
	"/v7/customers/signup": true,
}

func isGopayCustomerLinkPath(path string) bool {
	return path == "/v1/linkedapps" || strings.HasPrefix(path, "/v1/links/")
}

func isGopayCustomerAppHeaderPath(path string) bool {
	if gopayCustomerAppHeaderPaths[path] {
		return true
	}
	if path == "/v1/festivals" || strings.HasPrefix(path, "/v1/festivals/") {
		return true
	}
	if strings.HasPrefix(path, "/customers/v1/payments/") {
		return true
	}
	if strings.HasPrefix(path, "/v3/payments/") && strings.HasSuffix(path, "/capture") {
		return true
	}
	if strings.HasPrefix(path, "/api/v2/challenges/") && (strings.HasSuffix(path, "/pin-page") || strings.HasSuffix(path, "/pin-page/nb")) {
		return true
	}
	return false
}
