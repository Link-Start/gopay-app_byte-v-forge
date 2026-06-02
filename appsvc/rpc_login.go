package appsvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (s *Server) LoginStart(ctx context.Context, req *pb.LoginStartRequest) (*pb.LoginStartResponse, error) {
	state := s.parseRequestState(req.GetStateJson())
	if s.expireLoginIfNeeded(state) {
		// state mutated only; caller receives state_json below.
	}
	stage := stateString(state, "stage")
	if stage == "ready" || stage == "consumed" {
		tokenCheck := s.checkTokenValid(ctx, state)
		if s.tokenCheckReady(tokenCheck) {
			return &pb.LoginStartResponse{Success: true, OtpSent: false, StateJson: stateJSON(state)}, nil
		}
		if s.tokenCheckValid(tokenCheck) {
			return &pb.LoginStartResponse{Success: false, ErrorMessage: s.tokenCheckError(tokenCheck), StateJson: stateJSON(state)}, nil
		}
	}
	phone := strings.TrimSpace(req.GetPhone())
	if phone == "" && (stage == "login" || stage == "login_otp_pending") {
		phone = stateString(state, "_login_phone")
	}
	if phone == "" {
		return &pb.LoginStartResponse{Success: false, ErrorMessage: "login phone required", StateJson: stateJSON(state)}, nil
	}
	if stage == "login_otp_pending" && stateString(state, "_login_otp_token") != "" && stateString(state, "_login_verification_id") != "" {
		return &pb.LoginStartResponse{Success: true, OtpSent: true, VerificationId: stateString(state, "_login_verification_id"), VerificationMethod: stateString(state, "_login_verification_method"), StateJson: stateJSON(state)}, nil
	}
	result := s.startLogin(ctx, state, phone, req.GetPin(), req.GetCountryCode(), req.GetOtpChannel())
	if !anyBool(result["success"]) {
		errMessage := anyString(result["error"])
		if anyBool(result["not_registered"]) {
			errMessage = "GOPAY_PHONE_NOT_REGISTERED"
		}
		return &pb.LoginStartResponse{Success: false, ErrorMessage: errMessage, StateJson: stateJSON(state)}, nil
	}
	if anyBool(result["ready"]) {
		tokenCheck := s.checkTokenValid(ctx, state)
		if !s.tokenCheckReady(tokenCheck) {
			return &pb.LoginStartResponse{Success: false, ErrorMessage: s.tokenCheckError(tokenCheck), StateJson: stateJSON(state)}, nil
		}
	}
	return &pb.LoginStartResponse{
		Success:            true,
		OtpSent:            anyBool(result["otp_sent"]),
		VerificationId:     anyString(result["verification_id"]),
		VerificationMethod: anyString(result["method"]),
		StateJson:          stateJSON(state),
	}, nil
}

func (s *Server) LoginComplete(ctx context.Context, req *pb.LoginCompleteRequest) (*pb.LoginCompleteResponse, error) {
	state := s.parseRequestState(req.GetStateJson())
	s.expireLoginIfNeeded(state)
	if stateString(state, "stage") != "login_otp_pending" {
		return &pb.LoginCompleteResponse{Success: false, ErrorMessage: fmt.Sprintf("not waiting for login otp: %s", stringx.FirstNonEmpty(stateString(state, "stage"), "idle")), StateJson: stateJSON(state)}, nil
	}
	if err := s.completeLogin(ctx, state, req.GetOtp()); err != nil {
		return &pb.LoginCompleteResponse{Success: false, ErrorMessage: err.Error(), StateJson: stateJSON(state)}, nil
	}
	stage := stringx.FirstNonEmpty(stateString(state, "stage"), "idle")
	return &pb.LoginCompleteResponse{
		Success:            true,
		Phone:              stringx.FirstNonEmpty(stateString(state, "phone"), stateString(state, "_login_phone")),
		OtpSent:            stage == "login_otp_pending",
		VerificationId:     stateString(state, "_login_verification_id"),
		VerificationMethod: stateString(state, "_login_verification_method"),
		StateJson:          stateJSON(state),
	}, nil
}
