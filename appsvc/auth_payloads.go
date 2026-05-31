package appsvc

type signupProbeBody struct {
	PhoneNumber               string `json:"phone_number"`
	CountryCode               string `json:"country_code"`
	Email                     string `json:"email"`
	DeviceVerificationTokenID string `json:"device_verification_token_id"`
	ClientID                  string `json:"client_id"`
	ClientSecret              string `json:"client_secret"`
}

type signupMethodsBody struct {
	CountryCode               string `json:"country_code"`
	EmailAddress              any    `json:"email_address"`
	ClientID                  string `json:"client_id"`
	PhoneNumber               string `json:"phone_number"`
	ClientSecret              string `json:"client_secret"`
	Flow                      string `json:"flow"`
	DeviceVerificationTokenID any    `json:"device_verification_token_id"`
}

type signupInitiateBody struct {
	VerificationID            string `json:"verification_id"`
	Flow                      string `json:"flow"`
	VerificationMethod        string `json:"verification_method"`
	CountryCode               string `json:"country_code"`
	EmailAddress              any    `json:"email_address"`
	ClientID                  string `json:"client_id"`
	PhoneNumber               string `json:"phone_number"`
	ClientSecret              string `json:"client_secret"`
	IsMultipleMethod          any    `json:"is_multiple_method"`
	DeviceVerificationTokenID any    `json:"device_verification_token_id"`
}

type cvsVerifyBody struct {
	Data               any    `json:"data"`
	Flow               string `json:"flow"`
	VerificationID     string `json:"verification_id"`
	VerificationMethod string `json:"verification_method"`
	ClientID           string `json:"client_id"`
	ClientSecret       string `json:"client_secret"`
}

type gotoAuthClientBody struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type gotoCVSTokenBody struct {
	AccountID    string `json:"account_id"`
	ExtUserToken any    `json:"ext_user_token"`
	GrantType    string `json:"grant_type"`
	Token        string `json:"token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type gotoChallengeTokenBody struct {
	ExtUserToken any    `json:"ext_user_token"`
	GrantType    string `json:"grant_type"`
	Token        string `json:"token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}
