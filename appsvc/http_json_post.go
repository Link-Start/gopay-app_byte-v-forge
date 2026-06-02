package appsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/httpjson"
	"github.com/byte-v-forge/common-lib/jsonx"
	"github.com/byte-v-forge/common-lib/protojsonx"
	"google.golang.org/protobuf/proto"
)

var jsonPostSuccessStatuses = func() []int {
	statuses := make([]int, 0, http.StatusMultipleChoices-http.StatusOK)
	for status := http.StatusOK; status < http.StatusMultipleChoices; status++ {
		statuses = append(statuses, status)
	}
	return statuses
}()

type jsonPostOptions struct {
	Doer      httpjson.Doer
	Timeout   time.Duration
	BodyLimit int64
	Operation string
}

func postJSON(ctx context.Context, endpoint string, payload any, opts jsonPostOptions) error {
	_, err := doJSONPost(ctx, endpoint, payload, opts)
	return err
}

func doJSONPost(ctx context.Context, endpoint string, payload any, opts jsonPostOptions) (*httpjson.Response, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	raw, err := marshalPostJSONPayload(payload)
	if err != nil {
		return nil, err
	}
	options := []httpjson.Option{httpjson.WithBodyLimit(opts.BodyLimit)}
	if opts.Doer != nil {
		options = append(options, httpjson.WithHTTPDoer(opts.Doer))
	}
	client, err := httpjson.NewClient("", options...)
	if err != nil {
		return nil, err
	}
	return client.Do(reqCtx, httpjson.Request{
		Method:       http.MethodPost,
		Path:         endpoint,
		Body:         raw,
		Headers:      http.Header{"Content-Type": []string{"application/json"}},
		Operation:    strings.TrimSpace(opts.Operation),
		ExpectStatus: jsonPostSuccessStatuses,
	})
}

func marshalPostJSONPayload(payload any) ([]byte, error) {
	if message, ok := payload.(proto.Message); ok {
		return protojsonx.Marshal(message)
	}
	return jsonx.Compact(payload)
}
