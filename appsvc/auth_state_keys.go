package appsvc

const loginMethodsMaxAttempts = 3

var (
	loginStateKeys = []string{
		"_login_phone", "_login_country_code", "_login_verification_id",
		"_login_methods", "_login_default_method", "_login_methods_checked_at",
		"_login_flow", "_login_verification_method", "_login_otp_token", "_login_2fa_token",
		"_login_started_at", "_login_otp_sent_at", "_login_otp_expires_at",
	}
	signupAccountStateKeys = []string{"_signup_phone", "_signup_country_code", "_signup_name", "_signup_email"}
	signupOTPStateKeys     = []string{"_signup_verification_id", "_signup_verification_method", "_signup_otp_token", "_signup_started_at", "_signup_otp_sent_at", "_signup_otp_expires_at"}
	signupPINStateKeys     = []string{"_signup_pin_verification_id", "_signup_pin_verification_method", "_signup_pin_otp_token", "_signup_pin_challenge_id", "_signup_pin_client_id", "_signup_pin_otp_sent_at", "_signup_pin_otp_expires_at"}
	activeTokenKeys        = []string{"token", "refresh_token", "token_expires_at"}
	activeTokenMetaKeys    = []string{"last_token_refresh_at", "last_token_refresh_error", "last_token_refresh_failed_at"}
	tmpTokenKeys           = []string{"_tmp_token", "_tmp_refresh_token", "_tmp_token_expires_at"}
	tmpTokenMetaKeys       = []string{"_tmp_phone", "_tmp_token_migrated_at"}
)
