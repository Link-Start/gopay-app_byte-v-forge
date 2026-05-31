package paymentsvc

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"strings"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/byte-v-forge/common-lib/browserhttp"
	"github.com/byte-v-forge/common-lib/jsonx"
)

func (s *httpSession) request(ctx context.Context, method, rawURL string, opts requestOptions) (*httpResult, error) {
	var body io.Reader
	headers := cloneHeader(s.headers)
	mergeHeader(headers, opts.headers)
	if opts.jsonBody != nil {
		raw, err := jsonx.Compact(opts.jsonBody)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(raw)
		if headers.Get("Content-Type") == "" {
			headers.Set("Content-Type", "application/json")
		}
	} else if opts.formBody != nil {
		body = strings.NewReader(opts.formBody.Encode())
		headers.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if len(opts.query) > 0 {
		query := target.Query()
		for key, values := range opts.query {
			for _, value := range values {
				query.Add(key, value)
			}
		}
		target.RawQuery = query.Encode()
	}
	req, err := fhttp.NewRequestWithContext(ctx, strings.ToUpper(method), target.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header = browserhttp.ToFHTTPHeader(headers)
	client := s.client
	if opts.noRedirect {
		client.SetFollowRedirect(false)
		defer client.SetFollowRedirect(true)
	}
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
			if readErr != nil {
				return nil, readErr
			}
			payload, _ := jsonx.DecodeMap(raw)
			return &httpResult{status: resp.StatusCode, headers: browserhttp.FromFHTTPHeader(resp.Header), body: raw, json: map[string]any(payload)}, nil
		}
		lastErr = err
		if attempt >= 3 || !retryableTransportError(err) {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt) * time.Second):
		}
	}
	return nil, lastErr
}

func retryableTransportError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	for _, hint := range []string{"tls", "connection reset", "connection aborted", "timed out", "timeout", "temporarily unavailable", "network is unreachable", "proxy", "eof"} {
		if strings.Contains(text, hint) {
			return true
		}
	}
	return false
}
