package paymentsvc

import (
	"fmt"
	stdhttp "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/browserhttp"
)

const defaultTimeout = 30 * time.Second

type httpSession struct {
	client      browserhttp.TLSClient
	cookieJar   browserhttp.CookieJar
	proxyURL    string
	headers     stdhttp.Header
	fingerprint browserFingerprint
}

type requestOptions struct {
	headers    stdhttp.Header
	jsonBody   any
	formBody   url.Values
	query      url.Values
	noRedirect bool
}

func newHTTPSession(proxyURL string, fingerprints ...browserFingerprint) (*httpSession, error) {
	fingerprint := stablePaymentBrowserFingerprint(defaultBrowserLocale, "", "")
	if len(fingerprints) > 0 {
		fingerprint = fingerprints[0].withFallback(defaultBrowserLocale)
	}
	session := &httpSession{
		cookieJar:   browserhttp.NewCookieJar(),
		proxyURL:    strings.TrimSpace(proxyURL),
		headers:     make(stdhttp.Header),
		fingerprint: fingerprint,
	}
	if err := session.rebuildClient(fingerprint); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *httpSession) rebuildClient(fingerprint browserFingerprint) error {
	fingerprint = fingerprint.withFallback(defaultBrowserLocale)
	client, err := browserhttp.NewTLSClient(browserhttp.Config{
		Timeout:        defaultTimeout,
		ProxyURL:       s.proxyURL,
		TLSProfileName: fingerprint.TLSProfileName,
		DisableHTTP3:   true,
	}, s.cookieJar)
	if err != nil {
		return err
	}
	if s.client != nil {
		s.client.CloseIdleConnections()
	}
	s.client = client
	s.fingerprint = fingerprint
	return nil
}

func (s *httpSession) setProxy(proxyURL string) error {
	if s == nil {
		return fmt.Errorf("http session is nil")
	}
	proxyURL = strings.TrimSpace(proxyURL)
	if s.proxyURL == proxyURL {
		return nil
	}
	s.proxyURL = proxyURL
	return s.rebuildClient(s.fingerprint)
}

func (s *httpSession) close() {
	if s != nil && s.client != nil {
		s.client.CloseIdleConnections()
	}
}
