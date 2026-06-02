package appsvc

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/gopay-app/pb"
	"google.golang.org/protobuf/proto"
)

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,PUT,DELETE,OPTIONS")
		next.ServeHTTP(w, r)
	})
}

func noCacheFileServer(dir string) http.Handler {
	dir = strings.TrimSpace(dir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if dir == "" {
			http.NotFound(w, r)
			return
		}
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}
		http.NotFound(w, r)
	})
}

func writeProtoOrError(w http.ResponseWriter, value proto.Message, err error) {
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	_ = protojsonhttp.WriteResponse(w, http.StatusOK, value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	_ = protojsonhttp.WriteResponse(w, status, &pb.GopayAPIErrorResponse{Error: err.Error()})
}
