package app

import (
	"github.com/byte-v-forge/common-lib/envx"
	"github.com/byte-v-forge/common-lib/stringx"
)

type DeviceConfig struct {
	StaticIdentity   bool
	AppVersion       string
	AppID            string
	AppBuild         string
	AndroidVersion   string
	PhoneMake        string
	PhoneModel       string
	UniqueID         string
	SessionID        string
	TransactionID    string
	UserAgent        string
	D1               string
	XE2              string
	AdjTS            string
	AppsFlyerID      string
	WidevineID       string
	Screen           string
	WiFiMAC          string
	WiFiSSID         string
	M1ConnectionID   string
	M1Hardware       string
	M1Signature      string
	M1SignatureTime  string
	M1DeviceUUID     string
	FirebaseID       string
	AdvertisingID    string
	AppSetID         string
	InstallReferrer  string
	InstallerPackage string
	GMSVersion       string
	UserUUID         string
	DeviceToken      string
	IMEI             string
	IPAddress        string
	Location         string
	LocationAccuracy string
	GojekCountryCode string
	TLSProfileName   string
}

func DeviceConfigFromEnv() DeviceConfig {
	return DeviceConfig{
		StaticIdentity:   envx.Bool("GOPAY_STATIC_DEVICE_IDENTITY", false),
		AppVersion:       getenv("GOPAY_APP_VERSION"),
		AppID:            getenv("GOPAY_APP_ID"),
		AppBuild:         getenv("GOPAY_APP_BUILD"),
		AndroidVersion:   getenv("GOPAY_ANDROID_VERSION"),
		PhoneMake:        getenv("GOPAY_PHONE_MAKE"),
		PhoneModel:       getenv("GOPAY_PHONE_MODEL"),
		UniqueID:         getenv("GOPAY_UNIQUE_ID"),
		SessionID:        getenv("GOPAY_SESSION_ID"),
		TransactionID:    getenv("GOPAY_TRANSACTION_ID"),
		UserAgent:        getenv("GOPAY_USER_AGENT"),
		D1:               getenv("GOPAY_D1"),
		XE2:              getenv("GOPAY_X_E2"),
		AdjTS:            getenv("GOPAY_ADJ_TS"),
		AppsFlyerID:      getenv("GOPAY_APPSFLYER_ID"),
		WidevineID:       getenv("GOPAY_WIDEVINE_ID"),
		Screen:           getenv("GOPAY_SCREEN"),
		WiFiMAC:          getenv("GOPAY_WIFI_MAC"),
		WiFiSSID:         getenv("GOPAY_WIFI_SSID"),
		M1ConnectionID:   getenv("GOPAY_M1_CONNECTION_ID"),
		M1Hardware:       stringx.FirstNonEmpty(getenv("GOPAY_M1_HARDWARE"), getenv("GOPAY_M1_DEVICE_HARDWARE")),
		M1Signature:      getenv("GOPAY_M1_SIGNATURE"),
		M1SignatureTime:  getenv("GOPAY_M1_SIGNATURE_TIME"),
		M1DeviceUUID:     getenv("GOPAY_M1_DEVICE_UUID"),
		FirebaseID:       stringx.FirstNonEmpty(getenv("GOPAY_FIREBASE_APP_INSTANCE_ID"), getenv("GOPAY_FIREBASE_ID")),
		AdvertisingID:    stringx.FirstNonEmpty(getenv("GOPAY_ADVERTISING_ID"), getenv("GOPAY_AD_ID")),
		AppSetID:         getenv("GOPAY_APP_SET_ID"),
		InstallReferrer:  getenv("GOPAY_INSTALL_REFERRER"),
		InstallerPackage: getenv("GOPAY_INSTALLER_PACKAGE"),
		GMSVersion:       stringx.FirstNonEmpty(getenv("GOPAY_GMS_VERSION"), getenv("GOPAY_PLAY_SERVICES_VERSION")),
		UserUUID:         getenv("GOPAY_USER_UUID"),
		DeviceToken:      getenv("GOPAY_DEVICE_TOKEN"),
		IMEI:             getenv("GOPAY_IMEI"),
		IPAddress:        stringx.FirstNonEmpty(getenv("GOPAY_IP_ADDRESS"), getenv("GOPAY_LOCAL_IP_ADDRESS")),
		Location:         getenv("GOPAY_LOCATION"),
		LocationAccuracy: getenv("GOPAY_LOCATION_ACCURACY"),
		GojekCountryCode: getenv("GOPAY_COUNTRY_CODE"),
		TLSProfileName:   getenv("GOPAY_TLS_PROFILE"),
	}
}
