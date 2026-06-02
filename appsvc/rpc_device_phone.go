package appsvc

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) GenerateDeviceProxy(ctx context.Context, req *pb.GenerateDeviceProxyRequest) (*pb.GenerateDeviceProxyResponse, error) {
	if req == nil {
		req = &pb.GenerateDeviceProxyRequest{}
	}
	state, err := s.generateDeviceProxyState(ctx, req.GetAccountId(), req.GetCountryCode(), req.GetForceNew(), req.GetSkipPreflight(), req.GetEphemeralProfile())
	data := s.deviceProxyDiagnostics(state)
	if err != nil {
		return &pb.GenerateDeviceProxyResponse{Success: false, ErrorMessage: err.Error(), Data: protoStruct(data), StateJson: stateJSON(state)}, nil
	}
	return &pb.GenerateDeviceProxyResponse{
		Success:           true,
		ProxySlot:         int32(anyInt(data["proxy_slot"])),
		DynamicEgressSize: int32(anyInt(data["dynamic_egress_size"])),
		ProxyHash:         anyString(data["proxy_hash"]),
		DeviceFingerprint: anyString(data["device_fingerprint"]),
		Data:              protoStruct(data),
		StateJson:         stateJSON(state),
	}, nil
}

func protoStruct(data map[string]any) *structpb.Struct {
	if len(data) == 0 {
		return nil
	}
	value, err := structpb.NewStruct(data)
	if err != nil {
		return nil
	}
	return value
}

func (s *Server) CheckPhone(ctx context.Context, req *pb.CheckPhoneRequest) (*pb.CheckPhoneResponse, error) {
	phone := normalizePhoneWithConfig(s.cfg, req.GetPhone(), req.GetCountryCode())
	if phone == "" {
		return &pb.CheckPhoneResponse{Available: false, Status: "error", ErrorMessage: "phone required"}, nil
	}
	if strings.TrimSpace(req.GetStateJson()) == "" {
		return &pb.CheckPhoneResponse{Available: false, Status: "error", ErrorMessage: "generated device proxy state_json required"}, nil
	}
	proxyState := s.parseRequestState(req.GetStateJson())
	if stateString(proxyState, "_gopay_proxy") == "" {
		return &pb.CheckPhoneResponse{Available: false, Status: "error", ErrorMessage: "generated proxy missing", StateJson: stateJSON(proxyState)}, nil
	}
	if len(nestedMap(proxyState["device"])) == 0 {
		return &pb.CheckPhoneResponse{Available: false, Status: "error", ErrorMessage: "generated device missing", StateJson: stateJSON(proxyState)}, nil
	}
	result := s.checkPhoneByLoginMethods(ctx, phone, req.GetCountryCode(), proxyState)
	status := stringx.FirstNonEmpty(anyString(result["status"]), "error")
	errorMessage := anyString(result["error"])
	return &pb.CheckPhoneResponse{
		Available:         anyBool(result["available"]),
		Status:            status,
		ErrorMessage:      errorMessage,
		ProxySlot:         int32(anyInt(result["proxy_slot"])),
		DynamicEgressSize: int32(anyInt(result["dynamic_egress_size"])),
		ProxyHash:         anyString(result["proxy_hash"]),
		DeviceFingerprint: anyString(result["device_fingerprint"]),
		StateJson:         anyString(result["state_json"]),
	}, nil
}
