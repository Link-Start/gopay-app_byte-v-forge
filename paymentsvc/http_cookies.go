package paymentsvc

import (
	"net/url"
	"strings"
)

func (s *httpSession) cookieHeader(rawURL string) string {
	if s == nil || s.cookieJar == nil {
		return ""
	}
	target, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	cookies := s.cookieJar.Cookies(target)
	parts := make([]string, 0, len(cookies))
	seen := map[string]bool{}
	for _, cookie := range cookies {
		if cookie == nil || strings.TrimSpace(cookie.Name) == "" || cookie.Value == "" {
			continue
		}
		if seen[cookie.Name] {
			continue
		}
		seen[cookie.Name] = true
		parts = append(parts, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(parts, "; ")
}

func mergeCookieHeaders(values ...string) string {
	parts := make([]string, 0)
	seen := map[string]bool{}
	for _, value := range values {
		for _, raw := range strings.Split(value, ";") {
			part := strings.TrimSpace(raw)
			if part == "" || !strings.Contains(part, "=") {
				continue
			}
			name := strings.TrimSpace(strings.SplitN(part, "=", 2)[0])
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, "; ")
}
