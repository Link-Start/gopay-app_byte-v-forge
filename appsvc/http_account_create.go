package appsvc

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

func (h gopayHTTPHandler) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateGopayAccountRequest
	if err := protojsonhttp.ReadRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("request body must be json"))
		return
	}
	accountID, state, err := newGopayAccountInitialState(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.service.SaveGopayAccountState(r.Context(), &pb.SaveGopayAccountStateRequest{GopayAccountId: accountID, StateJson: stateJSON(state)})
	if err == nil && resp.GetSuccess() {
		_ = h.syncAccountProfileFields(r.Context(), accountID, stateString(state, "phone"), stateString(state, "_gopay_country_code"), "", stateString(state, "_otp_channel"), state)
	}
	writeProtoOrError(w, resp, err)
}

func newGopayAccountInitialState(req *pb.CreateGopayAccountRequest) (string, stateMap, error) {
	if strings.TrimSpace(req.GetCountryCode()) == "" {
		return "", nil, fmt.Errorf("country_code is required")
	}
	countryCode := phoneCountryCode(Config{}, req.GetCountryCode())
	if strings.HasPrefix(strings.TrimSpace(req.GetPhone()), "+") && !strings.HasPrefix(digitsOnly(req.GetPhone()), strings.TrimPrefix(countryCode, "+")) {
		return "", nil, fmt.Errorf("phone country code does not match country_code")
	}
	phone := normalizePhone(req.GetPhone(), countryCode)
	phone = strings.TrimLeft(phone, "0")
	if phone == "" {
		return "", nil, fmt.Errorf("phone is required")
	}
	otpChannel := normalizeActionOTPChannel(stringx.FirstNonEmpty(req.GetOtpChannel(), "wa"))
	if otpChannel == "" {
		otpChannel = "wa"
	}
	accountID, err := NormalizeGopayAccountID(newGoPayAccountID())
	if err != nil {
		return "", nil, err
	}
	now := time.Now().Unix()
	state := stateMap{
		"phone":                phone,
		"stage":                "created",
		"_gopay_country_code":  countryCode,
		"_signup_phone":        phone,
		"_signup_country_code": countryCode,
		"_login_phone":         phone,
		"_login_country_code":  countryCode,
		"_otp_channel":         otpChannel,
		"created_at_unix":      now,
		"updated_at_unix":      now,
	}
	persistGopayAccountOTPChannel(state)
	return accountID, state, nil
}

func newGoPayAccountID() string {
	return "gopay_" + randomProfileID()
}
