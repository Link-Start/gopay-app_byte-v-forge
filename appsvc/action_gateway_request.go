package appsvc

import "net/http"

func readGopayActionRequest(w http.ResponseWriter, r *http.Request) (gopayActionRequest, error) {
	var req gopayActionRequest
	if err := readOptionalActionJSONBody(w, r, 1<<20, &req); err != nil {
		return req, err
	}
	return req, nil
}
