package app

import (
	"context"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/gopay-app/protocol"
)

type session struct {
	token     string
	device    DeviceFingerprint
	http      *httpjson.Client
	signer    Signer
	logger    httpjson.Logger
	debugHTTP bool
}

func newSession(cfg Config) (*session, error) {
	device := cfg.Device
	if device.AppID == "" {
		var err error
		device, err = NewDeviceFingerprint(cfg.DeviceConfig)
		if err != nil {
			return nil, err
		}
	} else if device.TLSProfileName == "" {
		device.TLSProfileName = protocol.ResolveTLSProfileName("")
	}
	var httpClient httpjson.Doer
	if cfg.HTTPClient != nil {
		httpClient = cfg.HTTPClient
	} else {
		var err error
		httpClient, err = protocol.NewBrowserHTTPClient(cfg.Timeout, cfg.ProxyURL, device.TLSProfileName)
		if err != nil {
			return nil, err
		}
	}
	base, err := httpjson.NewClient("", httpjson.WithHTTPDoer(httpClient), httpjson.WithRetry(httpjson.RetryPolicy{Attempts: 1}), httpjson.WithLogger(cfg.Logger))
	if err != nil {
		return nil, err
	}
	return &session{
		token:  strings.TrimSpace(cfg.Token),
		device: device,
		http:   base,
		signer: Signer{
			SignVersion:           cfg.SignVersion,
			LegacyHMACKey:         cfg.LegacyHMACKey,
			DisplayEncoderKey:     cfg.DisplayEncoderKey,
			DisplayEncoderID:      cfg.DisplayEncoderID,
			SignedMsgTemplatePath: cfg.SignedMsgTemplatePath,
		},
		logger:    cfg.Logger,
		debugHTTP: cfg.DebugHTTP,
	}, nil
}

func (s *session) do(ctx context.Context, operation string, method string, rawURL string, body []byte, headers http.Header, expected ...int) (*httpjson.Response, error) {
	s.logHTTPRequest(ctx, method, rawURL, headers, body)
	resp, err := s.http.Do(ctx, httpjson.Request{
		Method:       method,
		Path:         rawURL,
		Body:         body,
		Headers:      headers,
		Operation:    operation,
		ExpectStatus: expected,
	})
	s.logHTTPResponse(ctx, method, rawURL, resp, err)
	return resp, err
}

func (s *session) signedHeaders(method string, rawURL string, body []byte, headers http.Header, extra http.Header, includeBodyMD5 bool) (http.Header, error) {
	if s.token != "" {
		if strings.HasPrefix(s.token, "Bearer ") {
			setHeader(headers, "Authorization", s.token)
		} else {
			setHeader(headers, "Authorization", "Bearer "+s.token)
		}
	}
	mergeHeaderValues(headers, extra)
	setRequestHost(headers, rawURL)
	signToken := headers.Get("Authorization")
	if signToken == "" {
		signToken = s.token
	}
	signature, err := s.signer.Sign(method, rawURL, body, signToken, s.device, s.device.XM1())
	if err != nil {
		return nil, err
	}
	setHeader(headers, "X-E1", signature.XE1)
	if includeBodyMD5 && s.signer.signVersionForRequest(rawURL) != "v2" {
		setHeader(headers, "X-E3", signature.BodyMD5)
	}
	return headers, nil
}
