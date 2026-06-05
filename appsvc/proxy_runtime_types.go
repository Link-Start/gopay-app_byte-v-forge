package appsvc

const (
	goPayProxyPurpose              = "gopay_app"
	goPayProxyCountryCode          = "ID"
	goPayProxyLeaseTTL             = "600s"
	goPayProxyProbeLeaseTTL        = "60s"
	goPayProxyPreflightMaxAttempts = 10
)

var goPayProxyConnectivityTargets = []string{
	"https://accounts.goto-products.com/",
	"https://customer.gopayapi.com/",
	"https://api.gojekapi.com/",
}

type proxyRuntimeAcquireOptions struct {
	AccountID     string
	CountryCode   string
	ForceNew      bool
	SkipPreflight bool
	LeaseTTL      string
}
