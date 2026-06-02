package appsvc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	digitsRE  = regexp.MustCompile(`\D+`)
	hexSeedRE = regexp.MustCompile(`[^0-9a-f]`)
)

func signupSeed(phone string) string {
	digits := digitsRE.ReplaceAllString(phone, "")
	if len(digits) > 6 {
		digits = digits[len(digits)-6:]
	}
	return fmt.Sprintf("%s%d", digits, time.Now().Unix())
}

func signupNameFromSeed(seed string) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	hexChars := hexSeedRE.ReplaceAllString(strings.ToLower(seed), "")
	if len(hexChars) < 2 {
		hexChars = fmt.Sprintf("%02s", hexChars)
	}
	hexChars = hexChars[len(hexChars)-2:]
	var out strings.Builder
	for _, ch := range hexChars {
		idx, _ := strconv.ParseInt(string(ch), 16, 64)
		out.WriteByte(alphabet[idx%int64(len(alphabet))])
	}
	return out.String()
}
