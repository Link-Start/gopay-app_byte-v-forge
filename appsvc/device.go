package appsvc

import (
	"reflect"
	"strings"

	gopayapp "github.com/byte-v-forge/gopay-app/protocol/app"
)

type deviceFieldMapping struct {
	Field   string
	Key     string
	Aliases []string
}

var deviceFieldMappings = []deviceFieldMapping{
	{Field: "AppType", Key: "x-apptype", Aliases: []string{"AppType"}},
	{Field: "AppVersion", Key: "x-appversion", Aliases: []string{"AppVersion"}},
	{Field: "AppID", Key: "x-appid", Aliases: []string{"AppID"}},
	{Field: "Platform", Key: "x-platform", Aliases: []string{"Platform"}},
	{Field: "UniqueID", Key: "x-uniqueid", Aliases: []string{"UniqueID"}},
	{Field: "PhoneMake", Key: "x-phonemake", Aliases: []string{"PhoneMake"}},
	{Field: "PhoneModel", Key: "x-phonemodel", Aliases: []string{"PhoneModel"}},
	{Field: "DeviceOS", Key: "x-deviceos", Aliases: []string{"DeviceOS"}},
	{Field: "UserType", Key: "x-user-type", Aliases: []string{"UserType"}},
	{Field: "SessionID", Key: "x-session-id", Aliases: []string{"SessionID"}},
	{Field: "TransactionID", Key: "transaction-id", Aliases: []string{"TransactionID"}},
	{Field: "UserAgent", Key: "user-agent", Aliases: []string{"UserAgent"}},
	{Field: "D1", Key: "d1", Aliases: []string{"D1"}},
	{Field: "XE2", Key: "x-e2", Aliases: []string{"XE2"}},
	{Field: "AdjTS", Key: "adjts", Aliases: []string{"AdjTS"}},
	{Field: "AppsFlyerID", Key: "m1_appsflyer_id", Aliases: []string{"AppsFlyerID"}},
	{Field: "WidevineID", Key: "m1_widevine_id", Aliases: []string{"WidevineID"}},
	{Field: "Screen", Key: "m1_screen", Aliases: []string{"Screen"}},
	{Field: "WiFiMAC", Key: "m1_wifi_mac", Aliases: []string{"WiFiMAC"}},
	{Field: "WiFiSSID", Key: "m1_wifi_ssid", Aliases: []string{"WiFiSSID"}},
	{Field: "M1ConnectionID", Key: "m1_connection_id", Aliases: []string{"M1ConnectionID"}},
	{Field: "M1Hardware", Key: "m1_hardware", Aliases: []string{"m1_device_hardware", "M1Hardware"}},
	{Field: "M1Signature", Key: "m1_signature", Aliases: []string{"M1Signature"}},
	{Field: "M1SignatureTime", Key: "m1_signature_time", Aliases: []string{"M1SignatureTime"}},
	{Field: "M1DeviceUUID", Key: "m1_device_uuid", Aliases: []string{"M1DeviceUUID"}},
	{Field: "FirebaseID", Key: "m1_firebase_app_instance_id", Aliases: []string{"firebase_app_instance_id", "FirebaseID"}},
	{Field: "AdvertisingID", Key: "advertising_id", Aliases: []string{"ad_id", "AdvertisingID"}},
	{Field: "AppSetID", Key: "app_set_id", Aliases: []string{"AppSetID"}},
	{Field: "InstallReferrer", Key: "install_referrer", Aliases: []string{"InstallReferrer"}},
	{Field: "InstallerPackage", Key: "installer_package", Aliases: []string{"InstallerPackage"}},
	{Field: "GMSVersion", Key: "gms_version", Aliases: []string{"play_services_version", "GMSVersion"}},
	{Field: "UserUUID", Key: "user-uuid", Aliases: []string{"UserUUID"}},
	{Field: "DeviceToken", Key: "x-devicetoken", Aliases: []string{"DeviceToken"}},
	{Field: "IMEI", Key: "x-imei", Aliases: []string{"IMEI"}},
	{Field: "IPAddress", Key: "x-ipaddress", Aliases: []string{"x-ip-address", "IPAddress"}},
	{Field: "Location", Key: "x-location", Aliases: []string{"Location"}},
	{Field: "LocationAccuracy", Key: "x-location-accuracy", Aliases: []string{"LocationAccuracy"}},
	{Field: "GojekCountryCode", Key: "gojek-country-code", Aliases: []string{"GojekCountryCode"}},
	{Field: "TLSProfileName", Key: "tls_profile", Aliases: []string{"tls-profile", "TLSProfileName"}},
}

func deviceFromMap(raw map[string]any) gopayapp.DeviceFingerprint {
	device := gopayapp.DeviceFingerprint{}
	dst := reflect.ValueOf(&device).Elem()
	for _, mapping := range deviceFieldMappings {
		field := dst.FieldByName(mapping.Field)
		if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.String {
			continue
		}
		if value := deviceMapString(raw, mapping); value != "" {
			field.SetString(value)
		}
	}
	return normalizeDeviceShape(device)
}

func deviceToMap(device gopayapp.DeviceFingerprint) map[string]any {
	src := reflect.ValueOf(device)
	out := make(map[string]any, len(deviceFieldMappings))
	for _, mapping := range deviceFieldMappings {
		field := src.FieldByName(mapping.Field)
		if !field.IsValid() || field.Kind() != reflect.String {
			out[mapping.Key] = ""
			continue
		}
		out[mapping.Key] = field.String()
	}
	return out
}

func mergeDevice(current, fallback gopayapp.DeviceFingerprint) gopayapp.DeviceFingerprint {
	out := current
	dst := reflect.ValueOf(&out).Elem()
	src := reflect.ValueOf(fallback)
	for i := 0; i < dst.NumField(); i++ {
		field := dst.Field(i)
		if field.Kind() != reflect.String || field.String() != "" {
			continue
		}
		fallbackField := src.Field(i)
		if fallbackField.Kind() == reflect.String {
			field.SetString(fallbackField.String())
		}
	}
	return normalizeDeviceShape(out)
}

func deviceMapString(raw map[string]any, mapping deviceFieldMapping) string {
	if raw == nil {
		return ""
	}
	if value := anyString(raw[mapping.Key]); value != "" {
		return value
	}
	for _, alias := range mapping.Aliases {
		if value := anyString(raw[alias]); value != "" {
			return value
		}
	}
	return ""
}

func normalizeDeviceShape(device gopayapp.DeviceFingerprint) gopayapp.DeviceFingerprint {
	device.PhoneModel = normalizeCommaValue(device.PhoneModel)
	device.DeviceOS = normalizeAndroidOSValue(device.DeviceOS)
	return device
}

func normalizeCommaValue(value string) string {
	value = strings.TrimSpace(value)
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return value
	}
	return strings.TrimSpace(parts[0]) + ", " + strings.TrimSpace(parts[1])
}

func normalizeAndroidOSValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(strings.ToLower(value), "android") {
		return value
	}
	return normalizeCommaValue(value)
}
