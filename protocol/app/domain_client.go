package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/common-lib/jsonx"
)

type domainClient struct {
	session   *session
	baseURL   string
	operation string
	headers   headerPolicy
}

type headerPolicy func(sess *session, method string, rawURL string, path string, body []byte, extra http.Header) (http.Header, error)

func newDomainClient(sess *session, baseURL string, operation string, headers headerPolicy) *domainClient {
	return &domainClient{session: sess, baseURL: baseURL, operation: operation, headers: headers}
}

func (c *domainClient) get(ctx context.Context, path string, expected ...int) (*httpjson.Response, error) {
	return c.request(ctx, http.MethodGet, path, nil, nil, expected...)
}

func (c *domainClient) post(ctx context.Context, path string, body any, expected ...int) (*httpjson.Response, error) {
	return c.request(ctx, http.MethodPost, path, body, nil, expected...)
}

func (c *domainClient) patch(ctx context.Context, path string, body any, expected ...int) (*httpjson.Response, error) {
	return c.request(ctx, http.MethodPatch, path, body, nil, expected...)
}

func (c *domainClient) put(ctx context.Context, path string, body any, expected ...int) (*httpjson.Response, error) {
	return c.request(ctx, http.MethodPut, path, body, nil, expected...)
}

func (c *domainClient) delete(ctx context.Context, path string, body any, expected ...int) (*httpjson.Response, error) {
	return c.request(ctx, http.MethodDelete, path, body, nil, expected...)
}

func (c *domainClient) request(ctx context.Context, method string, path string, body any, extra http.Header, expected ...int) (*httpjson.Response, error) {
	if c == nil || c.session == nil {
		return nil, fmt.Errorf("gopay domain client is nil")
	}
	rawURL, normalizedPath, err := c.endpoint(path)
	if err != nil {
		return nil, err
	}
	bodyRaw, err := jsonx.Compact(body)
	if err != nil {
		return nil, err
	}
	headers, err := c.headers(c.session, method, rawURL, normalizedPath, bodyRaw, extra)
	if err != nil {
		return nil, err
	}
	return c.session.do(ctx, c.operation, method, rawURL, bodyRaw, headers, expected...)
}

func (c *domainClient) endpoint(path string) (string, string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return c.sameDomainEndpoint(path)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	headerPath := path
	if ix := strings.IndexByte(headerPath, '?'); ix >= 0 {
		headerPath = headerPath[:ix]
	}
	return c.baseURL + path, headerPath, nil
}

func (c *domainClient) sameDomainEndpoint(rawURL string) (string, string, error) {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return "", "", err
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}
	if !strings.EqualFold(parsed.Scheme, base.Scheme) || !strings.EqualFold(parsed.Host, base.Host) {
		return "", "", fmt.Errorf("%s client refuses foreign host: %s", c.operation, parsed.Host)
	}
	headerPath := parsed.EscapedPath()
	if headerPath == "" {
		headerPath = "/"
	}
	return parsed.String(), headerPath, nil
}
