package app

import (
	"context"
	"net/http"

	"github.com/byte-v-forge/common-lib/httpjson"
)

type GojekClient struct{ domain *domainClient }

func (c *GojekClient) Get(ctx context.Context, path string, expected ...int) (*httpjson.Response, error) {
	return c.domain.get(ctx, path, expected...)
}

func (c *GojekClient) Post(ctx context.Context, path string, body any, expected ...int) (*httpjson.Response, error) {
	return c.domain.post(ctx, path, body, expected...)
}

func (c *GojekClient) Patch(ctx context.Context, path string, body any, expected ...int) (*httpjson.Response, error) {
	return c.domain.patch(ctx, path, body, expected...)
}

func (c *GojekClient) Request(ctx context.Context, method string, path string, body any, extra http.Header, expected ...int) (*httpjson.Response, error) {
	return c.domain.request(ctx, method, path, body, extra, expected...)
}

func gojekHeaderPolicy(sess *session, method string, rawURL string, path string, body []byte, extra http.Header) (http.Header, error) {
	hasBody := len(body) > 0
	headers := defaultSignedHeaders(sess.device, sess.device.XM1(), hasBody)
	if gojekActivityPaths[path] || gojekAppHeaderPaths[path] {
		headers = appHeaders(sess.device, sess.device.XM1(), hasBody)
	}
	return sess.signedHeaders(method, rawURL, body, headers, extra, true)
}
