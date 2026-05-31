package appsvc

import (
	"context"
	"strings"
	"time"

	"github.com/byte-v-forge/gopay-app/pb"
)

func (s *Server) GetGopayAccountProfile(ctx context.Context, req *pb.GetGopayAccountProfileRequest) (*pb.GetGopayAccountProfileResponse, error) {
	key, err := NormalizeGopayAccountID(req.GetGopayAccountId())
	if err != nil {
		return &pb.GetGopayAccountProfileResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	profile, err := s.loadGopayAccountProfile(ctx, key)
	if err != nil {
		return &pb.GetGopayAccountProfileResponse{Success: false, GopayAccountId: key, ErrorMessage: err.Error()}, nil
	}
	return gopayAccountProfileResponse(key, profile), nil
}

func (s *Server) SaveGopayAccountProfile(ctx context.Context, req *pb.SaveGopayAccountProfileRequest) (*pb.SaveGopayAccountProfileResponse, error) {
	key, err := NormalizeGopayAccountID(req.GetGopayAccountId())
	if err != nil {
		return &pb.SaveGopayAccountProfileResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	profile, err := s.loadGopayAccountProfile(ctx, key)
	if err != nil {
		return &pb.SaveGopayAccountProfileResponse{Success: false, GopayAccountId: key, ErrorMessage: err.Error()}, nil
	}
	if value := strings.TrimSpace(req.GetWaPhone()); value != "" {
		profile["wa_phone"] = value
	}
	if value := strings.TrimSpace(req.GetCountryCode()); value != "" {
		profile["country_code"] = value
	}
	if value := normalizeActionOTPChannel(req.GetOtpChannel()); value != "" {
		profile["otp_channel"] = value
	}
	if value := strings.TrimSpace(req.GetPin()); value != "" {
		profile["pin"] = value
	}
	profile["updated_at_unix"] = time.Now().Unix()
	if _, err := s.store.Save(ctx, gopayAccountProfileKey(key), stateJSON(profile)); err != nil {
		return &pb.SaveGopayAccountProfileResponse{Success: false, GopayAccountId: key, ErrorMessage: err.Error()}, nil
	}
	out := gopayAccountProfileResponse(key, profile)
	return &pb.SaveGopayAccountProfileResponse{
		Success:        out.GetSuccess(),
		ErrorMessage:   out.GetErrorMessage(),
		GopayAccountId: out.GetGopayAccountId(),
		WaPhone:        out.GetWaPhone(),
		CountryCode:    out.GetCountryCode(),
		UpdatedAtUnix:  out.GetUpdatedAtUnix(),
		OtpChannel:     out.GetOtpChannel(),
		PinConfigured:  out.GetPinConfigured(),
	}, nil
}

func (s *Server) loadGopayAccountProfile(ctx context.Context, gopayAccountID string) (stateMap, error) {
	raw, err := s.store.Load(ctx, gopayAccountProfileKey(gopayAccountID))
	if err != nil {
		return nil, err
	}
	return s.parseRequestState(raw), nil
}

func gopayAccountProfileResponse(gopayAccountID string, profile stateMap) *pb.GetGopayAccountProfileResponse {
	return &pb.GetGopayAccountProfileResponse{
		Success:        true,
		GopayAccountId: strings.TrimSpace(gopayAccountID),
		WaPhone:        stateString(profile, "wa_phone"),
		CountryCode:    stateString(profile, "country_code"),
		UpdatedAtUnix:  stateInt(profile, "updated_at_unix"),
		OtpChannel:     normalizeActionOTPChannel(stateString(profile, "otp_channel")),
		PinConfigured:  stateString(profile, "pin") != "",
	}
}

func gopayAccountProfileKey(gopayAccountID string) string {
	return "account-profile:" + strings.TrimSpace(gopayAccountID)
}
