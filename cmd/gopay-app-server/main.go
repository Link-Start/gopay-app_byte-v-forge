package main

import (
	"fmt"
	"log"
	"net"

	"github.com/byte-v-forge/common-lib/grpchealth"
	"github.com/byte-v-forge/gopay-app/appsvc"
	"github.com/byte-v-forge/gopay-app/pb"
	"google.golang.org/grpc"
)

func main() {
	cfg := appsvc.ConfigFromEnv()
	service, err := appsvc.NewServer(cfg)
	if err != nil {
		log.Fatalf("init gopay app service: %v", err)
	}
	defer func() { _ = service.Close() }()
	httpAPI, err := appsvc.StartHTTPAPI(cfg.HTTPListenAddr, cfg.DashboardStaticDir, cfg.N8NWebhookBaseURL, service)
	if err != nil {
		log.Fatalf("start gopay HTTP API: %v", err)
	}
	defer func() { _ = httpAPI.Close() }()
	otpWebhook, err := appsvc.StartOTPWebhook(cfg.OTPWebhookListenAddr, cfg.OTPSubmitURL)
	if err != nil {
		log.Fatalf("start gopay otp webhook: %v", err)
	}
	defer func() { _ = otpWebhook.Close() }()
	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("listen gopay app service: %v", err)
	}
	server := grpc.NewServer()
	pb.RegisterGopayAppServiceServer(server, service)
	grpchealth.RegisterServing(server)
	fmt.Printf("[gopay-app] Go gRPC listening on :%s\n", cfg.Port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve gopay app service: %v", err)
	}
}
