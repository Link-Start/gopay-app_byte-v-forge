package appsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/httpx"
	"github.com/byte-v-forge/common-lib/redactx"
)

var proxyRuntimeHTTPClient = &http.Client{Timeout: 35 * time.Second}

func postProxyRuntime(ctx context.Context, endpoint string, payload map[string]any, out any, timeout time.Duration) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := proxyRuntimeHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := httpx.ReadLimited(resp.Body, 1<<20)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d %s", resp.StatusCode, redactx.Snippet(redactx.Text(string(raw)), 300))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	return nil
}
