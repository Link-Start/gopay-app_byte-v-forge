FROM docker.m.daocloud.io/library/node:22-bookworm-slim AS dashboard_remote_builder

WORKDIR /gopay-app/webui
COPY common-lib/ui /common-lib/ui
COPY gopay-app/webui ./
RUN npm ci && SOURCE_ROOT=/ npm run build

FROM docker.m.daocloud.io/library/golang:1.26-alpine AS builder

WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
ENV PATH=/root/go/bin:$PATH

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache git protobuf-dev \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

COPY common-lib /common-lib
COPY gopay-app/go.mod gopay-app/go.sum ./
RUN go mod edit -replace github.com/byte-v-forge/common-lib=/common-lib \
    && go mod download

COPY gopay-app/proto ./proto
COPY gopay-app ./
RUN mkdir -p pb \
    && rm -f pb/*.pb.go pb/*_grpc.pb.go \
    && protoc -I proto -I /common-lib/proto --go_out=pb --go-grpc_out=pb proto/gopay_app.proto \
    && go build -o /out/gopay-app ./cmd/gopay-app-server

FROM docker.m.daocloud.io/library/alpine:latest

WORKDIR /app
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates

COPY --from=builder /out/gopay-app /app/bin/gopay-app
COPY --from=dashboard_remote_builder /gopay-app/webui/dist /app/dashboard/gopay

EXPOSE 50051 8080 8081
CMD ["/app/bin/gopay-app"]
