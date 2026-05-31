package app

const (
	defaultAppVersion      = "2.7.0"
	defaultAppID           = "com.gojek.gopay"
	defaultAppBuild        = "2070"
	defaultGojekCountry    = "ID"
	defaultAuthSDKVersion  = "1.0.0"
	defaultCVSDKVersion    = "1.0.0"
	defaultSupportSDK      = "0.44.0"
	defaultAcceptLanguage  = "en-ID"
	defaultTimezone        = "Asia/Jakarta"
	defaultUserLocale      = "en_ID"
	defaultAndroidVersion  = "7.0"
	defaultXE2             = "ED9A2B38749FBDE9ACA61D6A685B7"
	defaultPhoneMake       = "HUAWEI"
	defaultPhoneModel      = "HUAWEI, TRT-AL00A"
	defaultUniqueID        = "685b86605a047a3e"
	defaultD1              = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
	defaultAppsFlyerID     = "1779516675040-8955649077185556133"
	defaultWidevineID      = "T1B0eHZMQmFWV0h2UlBRSllIeVdlbFNtS1BqcXFiZwA="
	defaultXM1ConnectionID = "55093"
	defaultXM1Screen       = "720x1208"
	defaultXM1WiFiMAC      = "6c:b1:58:31:29:5b"
	defaultXM1WiFiSSID     = "Bug"
	defaultXM1Hardware     = "msm8937|1401|8"
	defaultM1Signature     = "0000000000000000"
	defaultM1DeviceUUID    = "00000000-0000-0000-0000-000000000000"
	defaultFirebaseID      = "00000000000000000000000000000000"
	defaultAdvertisingID   = "00000000-0000-0000-0000-000000000000"
	defaultAppSetID        = "00000000-0000-0000-0000-000000000000"
	defaultInstallReferrer = "utm_source=google-play&utm_medium=organic"
	defaultInstaller       = "com.android.vending"
	defaultGMSVersion      = "252014000"
	defaultLocation        = "-6.2000000,106.8000000"
	defaultLocationAcc     = "0.010999999552965164"
	defaultPlatform        = "Android"
	defaultUserType        = "customer"
	defaultApplicationType = "GOPAY"
)

type hardwareProfile struct {
	AndroidVersion string
	PhoneMake      string
	PhoneModel     string
	Screen         string
}

var hardwareProfiles = []hardwareProfile{
	{AndroidVersion: defaultAndroidVersion, PhoneMake: defaultPhoneMake, PhoneModel: defaultPhoneModel, Screen: defaultXM1Screen},
	{AndroidVersion: "12", PhoneMake: "samsung", PhoneModel: "samsung,SM-A525F", Screen: "1080x2174"},
	{AndroidVersion: "13", PhoneMake: "samsung", PhoneModel: "samsung,SM-A536E", Screen: "1080x2176"},
	{AndroidVersion: "13", PhoneMake: "samsung", PhoneModel: "samsung,SM-M336B", Screen: "1080x2193"},
	{AndroidVersion: "16", PhoneMake: "Xiaomi", PhoneModel: "Redmi,23117RK66C", Screen: "1080x2400"},
	{AndroidVersion: "13", PhoneMake: "Xiaomi", PhoneModel: "Redmi,2201117TY", Screen: "1080x2177"},
	{AndroidVersion: "12", PhoneMake: "Xiaomi", PhoneModel: "Redmi,M2101K7BNY", Screen: "1080x2150"},
	{AndroidVersion: "13", PhoneMake: "OPPO", PhoneModel: "OPPO,CPH2385", Screen: "1080x2172"},
	{AndroidVersion: "12", PhoneMake: "vivo", PhoneModel: "vivo,V2111", Screen: "1080x2179"},
}
