package app

import (
	"context"
	"net/http"

	"github.com/byte-v-forge/common-lib/httpjson"
)

type CustomerClient struct{ domain *domainClient }

func (c *CustomerClient) Get(ctx context.Context, path string, expected ...int) (*httpjson.Response, error) {
	return c.domain.get(ctx, path, expected...)
}

func (c *CustomerClient) Post(ctx context.Context, path string, body any, expected ...int) (*httpjson.Response, error) {
	return c.domain.post(ctx, path, body, expected...)
}

func (c *CustomerClient) Patch(ctx context.Context, path string, body any, expected ...int) (*httpjson.Response, error) {
	return c.domain.patch(ctx, path, body, expected...)
}

func (c *CustomerClient) Put(ctx context.Context, path string, body any, expected ...int) (*httpjson.Response, error) {
	return c.domain.put(ctx, path, body, expected...)
}

func (c *CustomerClient) Delete(ctx context.Context, path string, body any, expected ...int) (*httpjson.Response, error) {
	return c.domain.delete(ctx, path, body, expected...)
}

func (c *CustomerClient) Request(ctx context.Context, method string, path string, body any, extra http.Header, expected ...int) (*httpjson.Response, error) {
	return c.domain.request(ctx, method, path, body, extra, expected...)
}

func customerHeaderPolicy(sess *session, method string, rawURL string, path string, body []byte, extra http.Header) (http.Header, error) {
	hasBody := len(body) > 0
	var headers http.Header
	includeBodyMD5 := true
	switch {
	case path == "/v1/support/customer/initiate":
		headers = supportCustomerHeaders(sess.device, sess.device.XM1(), hasBody)
		includeBodyMD5 = false
	case isGopayCustomerLinkPath(path) || isGopayCustomerAppHeaderPath(path) || (method == http.MethodGet && gopayCustomerSlimGetPaths[path]):
		headers = appHeaders(sess.device, sess.device.XM1(), hasBody)
	default:
		headers = defaultSignedHeaders(sess.device, sess.device.XM1(), hasBody)
	}
	if path == "/api/v1/users/pin/tokens" {
		setHeader(headers, "Sdk-Version", sess.device.AppVersion)
		setHeader(headers, "X-Biometric", "")
		setHeader(headers, "X-Verification", "PIN")
	}
	return sess.signedHeaders(method, rawURL, body, headers, extra, includeBodyMD5)
}
