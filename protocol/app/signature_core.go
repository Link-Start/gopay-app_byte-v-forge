package app

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/stringx"
)

const emptyBodyMD5 = "d41d8cd98f00b204e9800998ecf8427e"

const (
	defaultGoPaySignVersion             = "auto"
	defaultGoPayDisplayEncoderID        = "D"
	defaultGoPayDisplayEncoderKey       = "1V79g&FZMB#zQ9:[T+8*xr1FXYVJ#%J)LiKl?c?=JG8dc{cX?d?p-u&Ti)$<vJC"
	defaultGoPayLegacyDisplayEncoderKey = "4&G6DbV&j8QZs~{)(Ila_w_|v@aqJq]E-;*(J9PanZ8sm01kTi{X<iG``]d7P&L"
	goPayV2TailConst                    = "c244dc56c7b6026a"
)

type Signer struct {
	Now                   func() time.Time
	SignVersion           string
	LegacyHMACKey         string
	DisplayEncoderKey     string
	DisplayEncoderID      string
	SignedMsgTemplatePath string
}

type Signature struct {
	XE1     string
	BodyMD5 string
}

func (s Signer) Sign(method string, rawURL string, body []byte, token string, device DeviceFingerprint, xM1 string) (Signature, error) {
	bodyMD5 := emptyBodyMD5
	if len(body) > 0 {
		sum := md5.Sum(body)
		bodyMD5 = hex.EncodeToString(sum[:])
	}
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	timestamp := fmt.Sprint(now.UnixMilli())
	version := s.signVersionForRequest(rawURL)
	if version == "v2" {
		return s.signV2(method, rawURL, token, device, xM1, timestamp, bodyMD5)
	}
	return s.signLegacy(method, rawURL, token, device, xM1, timestamp, bodyMD5)
}

func (s Signer) signVersionForRequest(rawURL string) string {
	configured := strings.ToLower(strings.TrimSpace(s.SignVersion))
	switch configured {
	case "v1", "legacy":
		return "v1"
	case "v2":
		return "v2"
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "v1"
	}
	host := strings.ToLower(parsed.Host)
	path := parsed.Path
	if host == "customer.gopayapi.com" {
		switch {
		case path == "/api/v1/users/pin/tokens/nb":
			return "v2"
		case path == "/v1/support/customer/activity":
			return "v2"
		}
	}
	return "v1"
}

func (s Signer) signLegacy(method string, rawURL string, token string, device DeviceFingerprint, xM1 string, timestamp string, bodyMD5 string) (Signature, error) {
	field1, err := randomField1()
	if err != nil {
		return Signature{}, err
	}
	path := signaturePath(rawURL)
	jwt := strings.TrimPrefix(token, "Bearer ")
	if xM1 == "" {
		xM1 = device.XM1()
	}
	parts := []string{
		device.AppType,
		device.PhoneModel + ":" + jwt,
		device.UniqueID + ":",
		bodyMD5 + ":" + path,
		strings.ToUpper(method) + ":" + timestamp,
		device.DeviceOS + ":" + device.AppVersion,
		xM1 + ":" + device.AppID,
		field1 + ":" + device.PhoneMake,
		device.Platform,
	}
	msg := strings.Join(parts, ";")
	mac := hmac.New(sha256.New, []byte(stringx.FirstNonEmpty(s.LegacyHMACKey, defaultGoPayLegacyDisplayEncoderKey)))
	_, _ = mac.Write([]byte(msg))
	return Signature{
		XE1:     hex.EncodeToString(mac.Sum(nil)) + ":" + field1 + ":D:" + timestamp,
		BodyMD5: bodyMD5,
	}, nil
}

func (s Signer) signV2(method string, rawURL string, token string, device DeviceFingerprint, xM1 string, timestamp string, bodyMD5 string) (Signature, error) {
	nonce, err := randomAlnum(32)
	if err != nil {
		return Signature{}, err
	}
	cipherHex, t3First16Hex := goPayV2Cipher(nonce)
	realMsg := s.goPayV2RealMsg(method, rawURL, token, device, xM1, timestamp, bodyMD5, cipherHex, t3First16Hex)
	shaHex := hmacSHA256Hex([]byte(stringx.FirstNonEmpty(s.DisplayEncoderKey, defaultGoPayDisplayEncoderKey)), realMsg)
	return Signature{
		XE1:     shaHex + ":" + cipherHex + ":" + stringx.FirstNonEmpty(s.DisplayEncoderID, defaultGoPayDisplayEncoderID) + ":" + timestamp,
		BodyMD5: bodyMD5,
	}, nil
}
