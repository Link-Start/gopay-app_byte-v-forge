package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/common-lib/jsonx"
	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/common-lib/timex"
)

type anyClient interface {
	Get(context.Context, string, ...int) (*httpjson.Response, error)
}

func (s *Server) loadGojekProfile(ctx context.Context, client anyClient) (map[string]any, string) {
	resp, err := client.Get(ctx, "/gojek/v2/customer")
	if err != nil {
		return nil, err.Error()
	}
	if resp.StatusCode != http.StatusOK {
		return nil, apiError("customer profile failed", resp)
	}
	for _, candidate := range []any{resp.Data(), resp.Payload} {
		if profile := gojekCustomerProfile(candidate); len(profile) > 0 {
			return profile, ""
		}
	}
	return nil, "customer profile missing"
}

func gojekCustomerProfile(value any) map[string]any {
	obj, ok := jsonx.Object(value)
	if !ok {
		return nil
	}
	for _, key := range []string{"customer", "profile", "user"} {
		if profile := nestedMap(obj[key]); len(profile) > 0 {
			return profile
		}
	}
	for _, key := range []string{"data", "raw"} {
		if profile := gojekCustomerProfile(obj[key]); len(profile) > 0 {
			return profile
		}
	}
	if looksLikeProfile(obj) {
		return obj
	}
	return nil
}

func looksLikeProfile(value map[string]any) bool {
	for _, key := range []string{"name", "email", "phone", "number", "profile_image_url", "profileImageUrl"} {
		if anyString(value[key]) != "" {
			return true
		}
	}
	return false
}

func (s *Server) syncProfileFields(state stateMap, profile map[string]any, countryCode string) {
	if name := anyString(profile["name"]); name != "" {
		state["name"] = name
	}
	if email := anyString(profile["email"]); email != "" {
		state["email"] = email
	}
	if phone := stringx.FirstNonEmpty(anyString(profile["phone"]), anyString(profile["number"])); phone != "" {
		state["phone"] = normalizePhone(phone, countryCode)
	}
}

func (s *Server) changePhoneProfileBody(state stateMap, countryCode, phone string) map[string]any {
	name := stringx.FirstNonEmpty(stateString(state, "name"), "gg")
	email := stateString(state, "email")
	if email == "" {
		email = fallbackProfileEmail(countryCode, phone)
	}
	state["name"] = name
	state["email"] = email
	return map[string]any{"email": email, "name": name, "phone": countryCode + phone, "profile_image_url": nil}
}

func fallbackProfileEmail(countryCode, phone string) string {
	digits := digitsRE.ReplaceAllString(countryCode+phone, "")
	if digits == "" {
		digits = fmt.Sprint(time.Now().Unix())
	}
	return "gopay" + digits + "@gmail.com"
}

func (s *Server) confirmChangePhone(ctx context.Context, client anyClient, countryCode, expectedPhone string) (bool, string) {
	expected := normalizePhone(expectedPhone, countryCode)
	deadline := time.Now().Add(s.cfg.ChangePhoneConfirmTimeout)
	last := ""
	for {
		profile, errMsg := s.loadGojekProfile(ctx, client)
		if errMsg != "" {
			last = errMsg
		} else {
			actual := normalizePhone(stringx.FirstNonEmpty(anyString(profile["phone"]), anyString(profile["number"])), countryCode)
			if actual == expected {
				return true, ""
			}
			last = fmt.Sprintf("phone change not confirmed: expected %s, got %s", expected, stringx.FirstNonEmpty(actual, "-"))
		}
		if time.Now().After(deadline) {
			return false, last
		}
		if err := timex.Sleep(ctx, s.cfg.ChangePhoneConfirmInterval); err != nil {
			return false, err.Error()
		}
	}
}

func phoneRegisteredResponse(resp *httpjson.Response) bool {
	if resp == nil || resp.StatusCode != http.StatusBadRequest {
		return false
	}
	text := strings.ToLower(responseText(resp))
	return strings.Contains(text, "user_can_not_update_phone") || strings.Contains(text, "already registered")
}
