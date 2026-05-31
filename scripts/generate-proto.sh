#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_ROOT="${SOURCE_ROOT:-$(cd "${ROOT}/.." && pwd)}"
COMMON_PROTO_DIR="${COMMON_PROTO_DIR:-${SOURCE_ROOT}/common-lib/proto}"
mkdir -p "${ROOT}/pb"
rm -f "${ROOT}/pb"/*.pb.go "${ROOT}/pb"/*_grpc.pb.go
protoc -I "${ROOT}/proto" -I "${COMMON_PROTO_DIR}" \
  --go_out="${ROOT}/pb" \
  --go-grpc_out="${ROOT}/pb" \
  "${ROOT}/proto/gopay_app.proto"
