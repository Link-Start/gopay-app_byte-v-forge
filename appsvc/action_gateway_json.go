package appsvc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func readOptionalActionJSONBody(w http.ResponseWriter, r *http.Request, limit int64, dst any) error {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, limit))
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	return json.Unmarshal(body, dst)
}

func writeActionJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
