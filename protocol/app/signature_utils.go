package app

import (
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/byte-v-forge/common-lib/randx"
)

func signaturePath(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return strings.TrimPrefix(strings.TrimPrefix(rawURL, "https://"), "http://")
	}
	return parsed.Host + parsed.RequestURI()
}

func randomAlnum(size int) (string, error) {
	return randx.String("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", size)
}

func randomField1() (string, error) {
	first, err := randx.Bytes(32)
	if err != nil {
		return "", err
	}
	middleA, err := randx.Bytes(2)
	if err != nil {
		return "", err
	}
	middleB, err := randx.Bytes(4)
	if err != nil {
		return "", err
	}
	second, err := randx.Bytes(16)
	if err != nil {
		return "", err
	}
	middle := "2000000040000000" +
		hex.EncodeToString(middleA) + "cf0f" +
		"28e4f5be08e4f5be" +
		hex.EncodeToString(middleB) +
		"c8e3f5befb1aad58"
	return hex.EncodeToString(first) + middle + hex.EncodeToString(second), nil
}

func lastHexRun(value []byte, size int) string {
	for idx := len(value) - size; idx >= 0; idx-- {
		candidate := value[idx : idx+size]
		if isHexASCII(candidate) {
			return string(candidate)
		}
	}
	return ""
}

func isHexASCII(value []byte) bool {
	for _, ch := range value {
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
			continue
		}
		return false
	}
	return true
}
