package app

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/httpjson"
)

type AuthClient struct{ domain *domainClient }

func (c *AuthClient) Get(ctx context.Context, path string, expected ...int) (*httpjson.Response, error) {
	return c.domain.get(ctx, path, expected...)
}

func (c *AuthClient) Post(ctx context.Context, path string, body any, expected ...int) (*httpjson.Response, error) {
	return c.domain.post(ctx, path, body, expected...)
}

func (c *AuthClient) Request(ctx context.Context, method string, path string, body any, extra http.Header, expected ...int) (*httpjson.Response, error) {
	return c.domain.request(ctx, method, path, body, extra, expected...)
}

func authHeaderPolicy(sess *session, method string, rawURL string, path string, body []byte, extra http.Header) (http.Header, error) {
	headers := gotoAuthHeaders(sess.device, sess.device.XM1(), len(body) > 0)
	if strings.HasPrefix(path, "/cvs/") {
		setHeader(headers, "Authorization", "")
	}
	if path == "/cvs/v1/initiate" && bytes.Contains(body, []byte(`"flow":"signup"`)) {
		setHeader(headers, "Key", "value")
	}
	return sess.signedHeaders(method, rawURL, body, headers, extra, false)
}
