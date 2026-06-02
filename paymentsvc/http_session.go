package paymentsvc

import (
	stdhttp "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/fingerprinthttp"
)

const defaultTimeout = 30 * time.Second

type httpSession struct {
	client      *fingerprinthttp.Client
	proxyURL    string
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
		proxyURL:    strings.TrimSpace(proxyURL),
		fingerprint: fingerprint,
	}
	if err := session.rebuildClient(fingerprint); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *httpSession) rebuildClient(fingerprint browserFingerprint) error {
	fingerprint = fingerprint.withFallback(defaultBrowserLocale)
	client, err := fingerprinthttp.New(fingerprinthttp.Config{
		Timeout:      defaultTimeout,
		ProxyURL:     s.proxyURL,
		Profile:      fingerprint.httpProfile(s.proxyURL),
		DisableHTTP3: true,
		RetryMax:     3,
		RetryDelay:   time.Second,
		MaxBodyBytes: 8 * 1024 * 1024,
	})
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

func (s *httpSession) close() {
	if s != nil && s.client != nil {
		s.client.CloseIdleConnections()
	}
}
