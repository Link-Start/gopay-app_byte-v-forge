package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/protojsonx"
	"google.golang.org/protobuf/proto"
)

var proxyRuntimeHTTPClient = &http.Client{Timeout: 35 * time.Second}

func postProxyRuntime(ctx context.Context, endpoint string, payload proto.Message, out proto.Message, timeout time.Duration) error {
	resp, err := doJSONPost(ctx, endpoint, payload, jsonPostOptions{
		Doer:      proxyRuntimeHTTPClient,
		Timeout:   timeout,
		BodyLimit: 1 << 20,
		Operation: "proxy runtime",
	})
	if err != nil || out == nil {
		return err
	}
	if err := protojsonx.Unmarshal(resp.Body, out); err != nil {
		return fmt.Errorf("parse proxy-runtime response: %w", err)
	}
	return nil
}
