package appsvc

import (
	"context"

	"github.com/byte-v-forge/gopay-app/paymentsvc"
	"github.com/byte-v-forge/gopay-app/pb"
)

type Server struct {
	pb.UnimplementedGopayAppServiceServer
	cfg     Config
	store   *StateStore
	payment *paymentsvc.Server
}

func NewServer(cfg Config) (*Server, error) {
	store, err := NewStateStore(context.Background(), cfg.StateRedisURL, cfg.StateKeyPrefix, cfg.StateTTL)
	if err != nil {
		return nil, err
	}
	return &Server{cfg: cfg, store: store, payment: paymentsvc.NewServer(cfg.Payment)}, nil
}

func (s *Server) Close() error {
	if s == nil || s.store == nil {
		return nil
	}
	return s.store.Close()
}
