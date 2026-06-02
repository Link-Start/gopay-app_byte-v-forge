package appsvc

import (
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/common-lib/randx"
)

func retryTimerSeconds(resp *httpjson.Response) int64 {
	if resp == nil {
		return 0
	}
	return maxRetryTimerSeconds(resp.Data()["retry_timer_in_seconds"])
}

func maxRetryTimerSeconds(value any) int64 {
	switch typed := value.(type) {
	case []any:
		var maxValue int64
		for _, item := range typed {
			if parsed := maxRetryTimerSeconds(item); parsed > maxValue {
				maxValue = parsed
			}
		}
		return maxValue
	default:
		return anyInt(value)
	}
}

func statusCode(resp *httpjson.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

func randomInt64(maxExclusive int64) int64 {
	if maxExclusive <= 0 {
		return 0
	}
	n, err := randx.Int(maxExclusive)
	if err != nil {
		return time.Now().UnixNano() % maxExclusive
	}
	return n
}

func rateLimitLabel(scope string) string {
	switch scope {
	case signupRateLimitScopeProbe:
		return "signup phone probe rate limited"
	case signupRateLimitScopeMethods:
		return "signup methods rate limited"
	case signupRateLimitScopeInitiate:
		return "signup otp initiate rate limited"
	default:
		return http.StatusText(http.StatusTooManyRequests)
	}
}
