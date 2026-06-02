package paymentsvc

import (
	"context"

	"github.com/byte-v-forge/common-lib/fingerprinthttp"
)

func (s *httpSession) request(ctx context.Context, method, rawURL string, opts requestOptions) (*httpResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	headers := cloneHeader(opts.headers)
	resp, err := s.client.Request(ctx, method, rawURL, fingerprinthttp.RequestOptions{
		Headers:    headers,
		JSONBody:   opts.jsonBody,
		FormBody:   opts.formBody,
		Query:      opts.query,
		NoRedirect: opts.noRedirect,
	})
	if err != nil {
		return nil, err
	}
	return &httpResult{status: resp.StatusCode, headers: resp.Headers, body: resp.Body, json: resp.JSON}, nil
}
