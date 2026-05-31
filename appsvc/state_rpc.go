package appsvc

import (
	"context"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (s *Server) GetGopayAccount(ctx context.Context, req *pb.GetGopayAccountRequest) (*pb.GetGopayAccountResponse, error) {
	key, err := NormalizeGopayAccountID(req.GetGopayAccountId())
	if err != nil {
		return &pb.GetGopayAccountResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	state, err := s.loadGopayAccountState(ctx, key)
	if err != nil {
		return &pb.GetGopayAccountResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	return &pb.GetGopayAccountResponse{Success: true, Account: gopayAccountProjection(key, state)}, nil
}

func (s *Server) ListGopayAccounts(ctx context.Context, req *pb.ListGopayAccountsRequest) (*pb.ListGopayAccountsResponse, error) {
	page, err := s.store.ListAccounts(ctx, req.GetCursor(), int(req.GetLimit()))
	if err != nil {
		return &pb.ListGopayAccountsResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	accounts := make([]*pb.GopayAccount, 0, len(page.Records))
	for _, record := range page.Records {
		state := s.parseRequestState(record.Raw)
		s.bindGoPayAccountIdentity(state, record.AccountID)
		accounts = append(accounts, gopayAccountProjection(record.AccountID, state))
	}
	return &pb.ListGopayAccountsResponse{Success: true, Accounts: accounts, NextCursor: page.NextCursor}, nil
}

func (s *Server) LoadGopayAccountState(ctx context.Context, req *pb.LoadGopayAccountStateRequest) (*pb.LoadGopayAccountStateResponse, error) {
	key, err := NormalizeGopayAccountID(req.GetGopayAccountId())
	if err != nil {
		return &pb.LoadGopayAccountStateResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	state, err := s.loadGopayAccountState(ctx, key)
	if err != nil {
		return &pb.LoadGopayAccountStateResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	return &pb.LoadGopayAccountStateResponse{Success: true, Account: gopayAccountProjection(key, state), StateJson: stateJSON(state)}, nil
}

func (s *Server) SaveGopayAccountState(ctx context.Context, req *pb.SaveGopayAccountStateRequest) (*pb.SaveGopayAccountStateResponse, error) {
	key, err := NormalizeGopayAccountID(req.GetGopayAccountId())
	if err != nil {
		return &pb.SaveGopayAccountStateResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	state := s.parseRequestState(stringx.FirstNonEmpty(req.GetStateJson(), "{}"))
	s.bindGoPayAccountIdentity(state, key)
	persistGopayAccountOTPChannel(state)
	if _, err := s.store.SaveAccount(ctx, key, stateJSON(state)); err != nil {
		return &pb.SaveGopayAccountStateResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	return &pb.SaveGopayAccountStateResponse{Success: true, Account: gopayAccountProjection(key, state)}, nil
}

func (s *Server) DeleteGopayAccount(ctx context.Context, req *pb.DeleteGopayAccountRequest) (*pb.DeleteGopayAccountResponse, error) {
	key, err := NormalizeGopayAccountID(req.GetGopayAccountId())
	if err != nil {
		return &pb.DeleteGopayAccountResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	if err := s.store.DeleteAccount(ctx, key); err != nil {
		return &pb.DeleteGopayAccountResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	return &pb.DeleteGopayAccountResponse{Success: true}, nil
}

func (s *Server) loadGopayAccountState(ctx context.Context, key string) (stateMap, error) {
	raw, err := s.store.LoadAccount(ctx, key)
	if err != nil {
		return nil, err
	}
	state := s.parseRequestState(raw)
	s.bindGoPayAccountIdentity(state, key)
	return state, nil
}
