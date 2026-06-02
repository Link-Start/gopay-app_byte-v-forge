package appsvc

import (
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/accountmodel"
	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/byte-v-forge/gopay-app/pb"
)

var gopayAccountDescriptor = accountmodel.Descriptor{SourceService: "gopay-app", AccountType: "gopay", ProviderKey: "gopay"}

func gopayAccountProjection(accountID string, state stateMap) *pb.GopayAccount {
	phone := stateString(state, "phone")
	countryCode := stateCountryCode(state)
	subjectPhone := gopayAccountPhoneE164(phone, countryCode)
	stage := stringx.FirstNonEmpty(stateString(state, "stage"), "idle")
	tokenPresent := stateString(state, "token") != ""
	pinSetup := anyBool(state["pin_setup"]) || stateInt(state, "pin_setup_at") > 0
	updatedAt := stateUpdatedAt(state)
	return &pb.GopayAccount{
		Account: gopayAccountDescriptor.Account(
			accountID,
			accountmodel.WithDisplayName(stringx.FirstNonEmpty(subjectPhone, phone, accountID)),
			accountmodel.WithSubject(accountmodel.PhoneSubject(subjectPhone, stringx.FirstNonEmpty(subjectPhone, phone))),
			accountmodel.WithStatus(accountmodel.StatusWithError(stage, accountmodel.StatusLabel(stage), "GOPAY_ACCOUNT_STATE_ERROR", stateString(state, "last_error"), true)),
			accountmodel.WithCredentials(
				accountmodel.Credential(accountmodel.CredentialKindToken, tokenPresent, gopayCredentialStatus(tokenPresent), stateExpiry(state, "token_expires_at"), updatedAt),
				accountmodel.Credential(accountmodel.CredentialKindPIN, pinSetup, gopayCredentialStatus(pinSetup), time.Time{}, updatedAt),
			),
			accountmodel.WithUpdatedAt(updatedAt),
		),
		Phone:           phone,
		CountryCode:     countryCode,
		BalanceAmount:   stateInt(state, "balance_amount"),
		BalanceCurrency: stateString(state, "balance_currency"),
		OtpChannel:      gopayAccountOTPChannelFromState(state),
	}
}

func gopayCredentialStatus(present bool) string {
	if !present {
		return ""
	}
	return accountmodel.CredentialStatusConfigured
}

func gopayAccountPhoneE164(phone string, countryCode string) string {
	value := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	if value == "" {
		return ""
	}
	prefix := strings.TrimPrefix(phoneCountryCode(Config{}, countryCode), "+")
	if strings.HasPrefix(value, prefix) {
		return "+" + value
	}
	return "+" + prefix + value
}

func stateCountryCode(state stateMap) string {
	return phoneCountryCode(Config{}, stringx.FirstNonEmpty(
		stateString(state, "_login_country_code"),
		stateString(state, "_signup_country_code"),
		stateString(state, "_gopay_country_code"),
	))
}

func stateUpdatedAt(state stateMap) time.Time {
	for _, key := range []string{"ready_at", "last_token_refresh_at", "last_balance_check_at", "pin_setup_checked_at"} {
		if value := stateInt(state, key); value > 0 {
			return time.Unix(value, 0).UTC()
		}
	}
	return time.Now().UTC()
}

func stateExpiry(state stateMap, key string) time.Time {
	if value := stateInt(state, key); value > 0 {
		return time.Unix(value, 0).UTC()
	}
	return time.Time{}
}
