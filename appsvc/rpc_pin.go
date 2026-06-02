package appsvc

import (
	"context"

	"github.com/byte-v-forge/gopay-app/pb"
)

func (s *Server) CreatePinStart(ctx context.Context, req *pb.CreatePinStartRequest) (*pb.CreatePinStartResponse, error) {
	state := s.parseRequestState(req.GetStateJson())
	s.expireSignupIfNeeded(state)
	result := s.startSignupPIN(ctx, state, req.GetPin(), req.GetOtpChannel())
	return &pb.CreatePinStartResponse{Success: anyBool(result["success"]), ErrorMessage: anyString(result["error"]), OtpSent: anyBool(result["otp_sent"]), VerificationId: anyString(result["verification_id"]), VerificationMethod: anyString(result["method"]), StateJson: stateJSON(state)}, nil
}

func (s *Server) CreatePinRetry(ctx context.Context, req *pb.CreatePinRetryRequest) (*pb.CreatePinRetryResponse, error) {
	state := s.parseRequestState(req.GetStateJson())
	s.expireSignupIfNeeded(state)
	result := s.retrySignupPIN(ctx, state)
	return &pb.CreatePinRetryResponse{Success: anyBool(result["success"]), ErrorMessage: anyString(result["error"]), OtpSent: anyBool(result["otp_sent"]), StateJson: stateJSON(state)}, nil
}

func (s *Server) CreatePinComplete(ctx context.Context, req *pb.CreatePinCompleteRequest) (*pb.CreatePinCompleteResponse, error) {
	state := s.parseRequestState(req.GetStateJson())
	s.expireSignupIfNeeded(state)
	result := s.completeSignupPIN(ctx, state, req.GetOtp(), req.GetPin())
	return &pb.CreatePinCompleteResponse{Success: anyBool(result["success"]), ErrorMessage: anyString(result["error"]), Phone: anyString(result["phone"]), PinSetupComplete: anyBool(result["pin_setup_complete"]), StateJson: stateJSON(state)}, nil
}
