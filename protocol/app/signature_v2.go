package app

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
)

func (s Signer) goPayV2RealMsg(method string, rawURL string, token string, device DeviceFingerprint, xM1 string, timestamp string, bodyMD5 string, cipherHex string, t3First16Hex string) []byte {
	key := []byte(stringx.FirstNonEmpty(s.DisplayEncoderKey, defaultGoPayDisplayEncoderKey))
	msg := goPayV2SyntheticRealMsg(method, rawURL, token, device, xM1, timestamp, bodyMD5, cipherHex, t3First16Hex, key)
	if path := strings.TrimSpace(s.SignedMsgTemplatePath); path != "" {
		if templ, err := os.ReadFile(path); err == nil && len(templ) > 0 {
			if patched := patchGoPayV2Template(templ, msg, token, cipherHex, t3First16Hex); len(patched) > 0 {
				return patched
			}
		}
	}
	return msg
}

func goPayV2SyntheticRealMsg(method string, rawURL string, token string, device DeviceFingerprint, xM1 string, timestamp string, bodyMD5 string, cipherHex string, t3First16Hex string, key []byte) []byte {
	jwt := strings.TrimPrefix(strings.TrimSpace(token), "Bearer ")
	if xM1 == "" {
		xM1 = device.XM1()
	}
	var msg bytes.Buffer
	msg.Write(hmacInnerPad(key))
	msg.WriteString(jwt)
	msg.WriteByte(':')
	msg.WriteString(stringx.FirstNonEmpty(device.PhoneModel, device.PhoneMake+", SM-G780F"))
	msg.WriteByte(':')
	msg.WriteString(stringx.FirstNonEmpty(xM1, device.XM1()))
	msg.WriteByte(':')
	msg.WriteString(device.AppVersion)
	msg.WriteByte(':')
	msg.WriteString(stringx.FirstNonEmpty(bodyMD5, emptyBodyMD5))
	msg.WriteByte(':')
	msg.WriteString(device.UniqueID)
	msg.WriteByte(':')
	msg.WriteString(strings.ToUpper(method))
	msg.WriteByte(':')
	msg.WriteString(device.DeviceOS)
	msg.WriteByte(':')
	msg.WriteString(timestamp)
	msg.WriteString("::")
	msg.WriteString(signaturePath(rawURL))
	msg.WriteByte(':')
	msg.WriteString(device.AppID)
	msg.WriteByte(':')
	msg.WriteString(cipherHex)
	msg.WriteString("0000000000000000")
	msg.WriteString("0000000000000000")
	msg.WriteString(goPayV2TailConst)
	msg.WriteString("0000000000000000")
	msg.WriteString(t3First16Hex)
	return msg.Bytes()
}

func patchGoPayV2Template(template []byte, fallback []byte, token string, cipherHex string, t3First16Hex string) []byte {
	patched := append([]byte(nil), template...)
	jwt := strings.TrimPrefix(strings.TrimSpace(token), "Bearer ")
	if jwt != "" {
		start := bytes.Index(patched, []byte("eyJhbGciOiJkaXIi"))
		end := bytes.Index(patched, []byte(":samsung,"))
		if start >= 0 && end > start {
			patched = append(append(append([]byte(nil), patched[:start]...), []byte(jwt)...), patched[end:]...)
		}
	}
	if len(patched) == 0 {
		return fallback
	}
	if idx := bytes.LastIndex(patched, []byte(goPayV2TailConst)); idx > 64 {
		searchStart := idx - 128
		if searchStart < 0 {
			searchStart = 0
		}
		window := patched[searchStart:idx]
		if old := lastHexRun(window, 64); old != "" {
			abs := searchStart + bytes.LastIndex(window, []byte(old))
			copy(patched[abs:abs+64], []byte(cipherHex))
		}
	}
	if len(t3First16Hex) == 32 && len(patched) >= 32 {
		copy(patched[len(patched)-32:], []byte(t3First16Hex))
	}
	return patched
}

func goPayV2Cipher(nonce string) (string, string) {
	zeroKey := make([]byte, 64)
	hkdfData := bytes.Repeat([]byte{1}, 64)
	expandTag := bytes.Repeat([]byte{1}, 32)
	keyC := hmacSHA256(zeroKey, hkdfData)
	keyD := hmacSHA256(keyC, hkdfData)
	k9Input := append(append(append(append([]byte{}, keyD...), expandTag...), keyC...), []byte(nonce)...)
	k9 := hmacSHA256(keyC, k9Input)
	t1 := hmacSHA256(k9, append(append([]byte{}, keyD...), expandTag...))
	t2 := hmacSHA256(k9, append(append([]byte{}, t1...), expandTag...))
	t3 := hmacSHA256(k9, append(append([]byte{}, t2...), expandTag...))
	return hex.EncodeToString(t2), hex.EncodeToString(t3[:16])
}

func hmacSHA256(key []byte, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}

func hmacSHA256Hex(key []byte, data []byte) string {
	return hex.EncodeToString(hmacSHA256(key, data))
}

func hmacInnerPad(key []byte) []byte {
	block := make([]byte, sha256.BlockSize)
	if len(key) > sha256.BlockSize {
		sum := sha256.Sum256(key)
		key = sum[:]
	}
	copy(block, key)
	for idx := range block {
		block[idx] ^= 0x36
	}
	return block
}
