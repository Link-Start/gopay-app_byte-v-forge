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
	defaultIOSVersion      = "18.5"
	defaultAndroidVersion  = "7.0"
	defaultXE2             = "ED9A2B38749FBDE9ACA61D6A685B7"
	defaultPhoneMake       = "Apple"
	defaultPhoneModel      = "Apple, iPhone16,2"
	defaultUniqueID        = "685b86605a047a3e"
	defaultD1              = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
	defaultAppsFlyerID     = "1779516675040-8955649077185556133"
	defaultWidevineID      = "T1B0eHZMQmFWV0h2UlBRSllIeVdlbFNtS1BqcXFiZwA="
	defaultXM1ConnectionID = "55093"
	defaultXM1Screen       = "1290x2796"
	defaultXM1WiFiMAC      = "6c:b1:58:31:29:5b"
	defaultXM1WiFiSSID     = "Bug"
	defaultXM1Hardware     = "iPhone16,2|6144|8"
	defaultM1Signature     = "0000000000000000"
	defaultM1DeviceUUID    = "00000000-0000-0000-0000-000000000000"
	defaultFirebaseID      = "00000000000000000000000000000000"
	defaultAdvertisingID   = "00000000-0000-0000-0000-000000000000"
	defaultAppSetID        = "00000000-0000-0000-0000-000000000000"
	defaultInstallReferrer = "utm_source=app-store&utm_medium=organic"
	defaultInstaller       = "com.apple.AppStore"
	defaultGMSVersion      = "252014000"
	defaultLocation        = "-6.2000000,106.8000000"
	defaultLocationAcc     = "0.010999999552965164"
	defaultPlatform        = "iOS"
	defaultUserType        = "customer"
	defaultApplicationType = "GOPAY"

	androidPlatform        = "Android"
	androidPhoneMake       = "HUAWEI"
	androidPhoneModel      = "HUAWEI, TRT-AL00A"
	androidXM1Hardware     = "msm8937|1401|8"
	androidXM1Screen       = "720x1208"
	androidInstallReferrer = "utm_source=google-play&utm_medium=organic"
	androidInstaller       = "com.android.vending"
)

type hardwareProfile struct {
	Platform   string
	OSVersion  string
	PhoneMake  string
	PhoneModel string
	Screen     string
	Hardware   string
}

var appleHardwareProfiles = []hardwareProfile{
	{Platform: defaultPlatform, OSVersion: defaultIOSVersion, PhoneMake: defaultPhoneMake, PhoneModel: defaultPhoneModel, Screen: defaultXM1Screen, Hardware: defaultXM1Hardware},
	{Platform: defaultPlatform, OSVersion: "18.5", PhoneMake: "Apple", PhoneModel: "Apple, iPhone16,1", Screen: "1179x2556", Hardware: "iPhone16,1|6144|8"},
	{Platform: defaultPlatform, OSVersion: "18.4", PhoneMake: "Apple", PhoneModel: "Apple, iPhone15,3", Screen: "1290x2796", Hardware: "iPhone15,3|6144|8"},
	{Platform: defaultPlatform, OSVersion: "17.7", PhoneMake: "Apple", PhoneModel: "Apple, iPhone14,5", Screen: "1170x2532", Hardware: "iPhone14,5|6144|8"},
}

var androidHardwareProfiles = []hardwareProfile{
	{Platform: androidPlatform, OSVersion: defaultAndroidVersion, PhoneMake: androidPhoneMake, PhoneModel: androidPhoneModel, Screen: androidXM1Screen, Hardware: androidXM1Hardware},
	{Platform: androidPlatform, OSVersion: "12", PhoneMake: "samsung", PhoneModel: "samsung,SM-A525F", Screen: "1080x2174", Hardware: "samsung|2800|8"},
	{Platform: androidPlatform, OSVersion: "13", PhoneMake: "samsung", PhoneModel: "samsung,SM-A536E", Screen: "1080x2176", Hardware: "samsung|2400|8"},
	{Platform: androidPlatform, OSVersion: "13", PhoneMake: "samsung", PhoneModel: "samsung,SM-M336B", Screen: "1080x2193", Hardware: "samsung|2400|8"},
	{Platform: androidPlatform, OSVersion: "16", PhoneMake: "Xiaomi", PhoneModel: "Redmi,23117RK66C", Screen: "1080x2400", Hardware: "qcom|3200|8"},
	{Platform: androidPlatform, OSVersion: "13", PhoneMake: "Xiaomi", PhoneModel: "Redmi,2201117TY", Screen: "1080x2177", Hardware: "qcom|2400|8"},
	{Platform: androidPlatform, OSVersion: "12", PhoneMake: "Xiaomi", PhoneModel: "Redmi,M2101K7BNY", Screen: "1080x2150", Hardware: "mtk|2400|8"},
	{Platform: androidPlatform, OSVersion: "13", PhoneMake: "OPPO", PhoneModel: "OPPO,CPH2385", Screen: "1080x2172", Hardware: "qcom|2400|8"},
	{Platform: androidPlatform, OSVersion: "12", PhoneMake: "vivo", PhoneModel: "vivo,V2111", Screen: "1080x2179", Hardware: "qcom|2400|8"},
}
