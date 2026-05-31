package app

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/protocol"
)

type DeviceFingerprint struct {
	AppType          string
	AppVersion       string
	AppID            string
	Platform         string
	UniqueID         string
	PhoneMake        string
	PhoneModel       string
	DeviceOS         string
	UserType         string
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

func NewDeviceFingerprint(cfg DeviceConfig) (DeviceFingerprint, error) {
	profile := randomHardwareProfile(cfg.StaticIdentity)
	appVersion := stringx.FirstNonEmpty(cfg.AppVersion, defaultAppVersion)
	appID := stringx.FirstNonEmpty(cfg.AppID, defaultAppID)
	appBuild := stringx.FirstNonEmpty(cfg.AppBuild, defaultAppBuild)
	deviceOS := androidDeviceOS(stringx.FirstNonEmpty(cfg.AndroidVersion, profile.AndroidVersion))
	phoneMake := stringx.FirstNonEmpty(cfg.PhoneMake, profile.PhoneMake)
	phoneModel := normalizePhoneModel(stringx.FirstNonEmpty(cfg.PhoneModel, profile.PhoneModel))
	userAgent := stringx.FirstNonEmpty(cfg.UserAgent, fmt.Sprintf("GoPay/%s (%s; build:%s; %s)", appVersion, appID, appBuild, deviceOS))
	uniqueID := stringx.FirstNonEmpty(cfg.UniqueID, generatedOrStatic(cfg.StaticIdentity, defaultUniqueID, func() string { return randomHex(8) }))
	d1 := stringx.FirstNonEmpty(cfg.D1, generatedOrStatic(cfg.StaticIdentity, defaultD1, randomD1))
	appsFlyerID := stringx.FirstNonEmpty(cfg.AppsFlyerID, generatedOrStatic(cfg.StaticIdentity, defaultAppsFlyerID, randomAppsFlyerID))
	widevineID := stringx.FirstNonEmpty(cfg.WidevineID, generatedOrStatic(cfg.StaticIdentity, defaultWidevineID, randomWidevineID))
	wifiMAC := strings.ToLower(stringx.FirstNonEmpty(cfg.WiFiMAC, generatedOrStatic(cfg.StaticIdentity, defaultXM1WiFiMAC, randomWiFiMAC)))
	wifiSSID := stringx.FirstNonEmpty(cfg.WiFiSSID, generatedOrStatic(cfg.StaticIdentity, defaultXM1WiFiSSID, randomWiFiSSID))
	m1ConnectionID := stringx.FirstNonEmpty(cfg.M1ConnectionID, generatedOrStatic(cfg.StaticIdentity, defaultXM1ConnectionID, randomM1ConnectionID))
	m1Hardware := stringx.FirstNonEmpty(cfg.M1Hardware, defaultXM1Hardware)
	m1Signature := stringx.FirstNonEmpty(cfg.M1Signature, generatedOrStatic(cfg.StaticIdentity, defaultM1Signature, func() string { return randomHex(8) }))
	m1SignatureTime := stringx.FirstNonEmpty(cfg.M1SignatureTime, generatedOrStatic(cfg.StaticIdentity, "0", randomM1SignatureTime))
	m1DeviceUUID := stringx.FirstNonEmpty(cfg.M1DeviceUUID, generatedOrStatic(cfg.StaticIdentity, defaultM1DeviceUUID, uuid.NewString))
	firebaseID := stringx.FirstNonEmpty(cfg.FirebaseID, generatedOrStatic(cfg.StaticIdentity, defaultFirebaseID, func() string { return randomHex(16) }))
	advertisingID := stringx.FirstNonEmpty(cfg.AdvertisingID, generatedOrStatic(cfg.StaticIdentity, defaultAdvertisingID, uuid.NewString))
	appSetID := stringx.FirstNonEmpty(cfg.AppSetID, generatedOrStatic(cfg.StaticIdentity, defaultAppSetID, uuid.NewString))
	deviceToken := strings.TrimSpace(cfg.DeviceToken)
	return DeviceFingerprint{
		AppType:          defaultApplicationType,
		AppVersion:       appVersion,
		AppID:            appID,
		Platform:         defaultPlatform,
		UniqueID:         uniqueID,
		PhoneMake:        phoneMake,
		PhoneModel:       phoneModel,
		DeviceOS:         deviceOS,
		UserType:         defaultUserType,
		SessionID:        stringx.FirstNonEmpty(cfg.SessionID, uuid.NewString()),
		TransactionID:    stringx.FirstNonEmpty(cfg.TransactionID, uuid.NewString()),
		UserAgent:        userAgent,
		D1:               d1,
		XE2:              stringx.FirstNonEmpty(cfg.XE2, defaultXE2),
		AdjTS:            stringx.FirstNonEmpty(cfg.AdjTS, "host:D"),
		AppsFlyerID:      appsFlyerID,
		WidevineID:       widevineID,
		Screen:           stringx.FirstNonEmpty(cfg.Screen, profile.Screen),
		WiFiMAC:          wifiMAC,
		WiFiSSID:         wifiSSID,
		M1ConnectionID:   m1ConnectionID,
		M1Hardware:       m1Hardware,
		M1Signature:      m1Signature,
		M1SignatureTime:  m1SignatureTime,
		M1DeviceUUID:     m1DeviceUUID,
		FirebaseID:       firebaseID,
		AdvertisingID:    advertisingID,
		AppSetID:         appSetID,
		InstallReferrer:  stringx.FirstNonEmpty(cfg.InstallReferrer, defaultInstallReferrer),
		InstallerPackage: stringx.FirstNonEmpty(cfg.InstallerPackage, defaultInstaller),
		GMSVersion:       stringx.FirstNonEmpty(cfg.GMSVersion, defaultGMSVersion),
		UserUUID:         strings.TrimSpace(cfg.UserUUID),
		DeviceToken:      deviceToken,
		IMEI:             stringx.FirstNonEmpty(cfg.IMEI, uniqueID),
		IPAddress:        stringx.FirstNonEmpty(cfg.IPAddress, generatedOrStatic(cfg.StaticIdentity, "", randomPrivateIP)),
		Location:         stringx.FirstNonEmpty(cfg.Location, defaultLocation),
		LocationAccuracy: stringx.FirstNonEmpty(cfg.LocationAccuracy, defaultLocationAcc),
		GojekCountryCode: stringx.FirstNonEmpty(cfg.GojekCountryCode, defaultGojekCountry),
		TLSProfileName:   protocol.ResolveTLSProfileName(cfg.TLSProfileName),
	}, nil
}

func (d DeviceFingerprint) XM1() string {
	if d.usesGoPay27Profile() {
		return strings.Join([]string{
			"3:" + stringx.FirstNonEmpty(d.AppsFlyerID, defaultAppsFlyerID),
			"4:" + stringx.FirstNonEmpty(d.M1ConnectionID, defaultXM1ConnectionID),
			"5:" + stringx.FirstNonEmpty(d.M1Hardware, defaultXM1Hardware),
			"6:" + stringx.FirstNonEmpty(d.WiFiMAC, defaultXM1WiFiMAC),
			"7:" + stringx.FirstNonEmpty(d.WiFiSSID, defaultXM1WiFiSSID),
			"8:" + stringx.FirstNonEmpty(d.Screen, defaultXM1Screen),
			"10:0",
			"11:" + stringx.FirstNonEmpty(d.WidevineID, defaultWidevineID),
			"15:" + stringx.FirstNonEmpty(d.FirebaseID, defaultFirebaseID),
		}, ",")
	}
	return strings.Join([]string{
		"3:" + stringx.FirstNonEmpty(d.AppsFlyerID, defaultAppsFlyerID),
		"4:" + stringx.FirstNonEmpty(d.M1ConnectionID, defaultXM1ConnectionID),
		"5:" + stringx.FirstNonEmpty(d.PhoneMake, defaultPhoneMake) + "|3200|2",
		"6:" + stringx.FirstNonEmpty(d.WiFiMAC, defaultXM1WiFiMAC),
		"7:" + stringx.FirstNonEmpty(d.WiFiSSID, defaultXM1WiFiSSID),
		"8:" + stringx.FirstNonEmpty(d.Screen, defaultXM1Screen),
		"9:passive,network,fused,gps",
		"10:1",
		"11:" + stringx.FirstNonEmpty(d.WidevineID, defaultWidevineID),
		"13:" + stringx.FirstNonEmpty(d.M1Signature, defaultM1Signature),
		"14:" + stringx.FirstNonEmpty(d.M1SignatureTime, "0"),
		"15:" + stringx.FirstNonEmpty(d.FirebaseID, defaultFirebaseID),
		"16:" + stringx.FirstNonEmpty(d.M1DeviceUUID, defaultM1DeviceUUID),
	}, ",")
}

func (d DeviceFingerprint) usesGoPay27Profile() bool {
	version := strings.TrimSpace(d.AppVersion)
	return version == "" || version == "2.7" || strings.HasPrefix(version, "2.7.")
}

func (d DeviceFingerprint) WithNewTransactionID() DeviceFingerprint {
	d.TransactionID = uuid.NewString()
	return d
}
