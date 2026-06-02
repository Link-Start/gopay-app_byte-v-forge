package appsvc

import (
	"context"
	"strings"

	"github.com/byte-v-forge/gopay-app/pb"
)

func (s *Server) SignupStart(ctx context.Context, req *pb.SignupStartRequest) (*pb.SignupStartResponse, error) {
	state := s.parseRequestState(req.GetStateJson())
	s.expireLoginIfNeeded(state)
	s.expireSignupIfNeeded(state)
	if strings.TrimSpace(req.GetPhone()) == "" {
		return &pb.SignupStartResponse{Success: false, ErrorMessage: "signup phone required", StateJson: stateJSON(state)}, nil
	}
	result := s.startSignup(ctx, state, req.GetPhone(), req.GetName(), req.GetEmail(), req.GetCountryCode(), req.GetOtpChannel(), req.GetSkipPhoneProbe())
	if !anyBool(result["success"]) {
		return &pb.SignupStartResponse{Success: false, ErrorMessage: anyString(result["error"]), RawJson: anyString(result["raw_json"]), StateJson: stateJSON(state)}, nil
	}
	return &pb.SignupStartResponse{
		Success:            true,
		OtpSent:            anyBool(result["otp_sent"]),
		VerificationId:     anyString(result["verification_id"]),
		VerificationMethod: anyString(result["method"]),
		RawJson:            anyString(result["raw_json"]),
		RetryTimerSeconds:  int32Slice(result["retry_timer_seconds"]),
		StateJson:          stateJSON(state),
	}, nil
}

func (s *Server) SignupRetry(ctx context.Context, req *pb.SignupRetryRequest) (*pb.SignupRetryResponse, error) {
	state := s.parseRequestState(req.GetStateJson())
	result := s.retrySignupOTP(ctx, state)
	return &pb.SignupRetryResponse{Success: anyBool(result["success"]), ErrorMessage: anyString(result["error"]), OtpSent: anyBool(result["otp_sent"]), RawJson: anyString(result["raw_json"]), StateJson: stateJSON(state)}, nil
}

func (s *Server) SignupComplete(ctx context.Context, req *pb.SignupCompleteRequest) (*pb.SignupCompleteResponse, error) {
	state := s.parseRequestState(req.GetStateJson())
	s.expireSignupIfNeeded(state)
	result := s.completeSignup(ctx, state, req.GetOtp())
	return &pb.SignupCompleteResponse{Success: anyBool(result["success"]), ErrorMessage: anyString(result["error"]), Phone: anyString(result["phone"]), PinSetupRequired: anyBool(result["pin_setup_required"]), RawJson: anyString(result["raw_json"]), StateJson: stateJSON(state)}, nil
}
