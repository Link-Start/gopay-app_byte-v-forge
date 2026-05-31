package appsvc

const (
	goPayProxyPurpose              = "gopay_app"
	goPayProxyCountryCode          = "ID"
	goPayProxyLeaseTTL             = "600s"
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
}

type proxyRuntimeLeaseResponse struct {
	Lease struct {
		LeaseID           string               `json:"lease_id"`
		ProviderAccountID string               `json:"provider_account_id"`
		ExpiresAt         string               `json:"expires_at"`
		Session           proxyRuntimeSession  `json:"session"`
		Egress            proxyRuntimeEndpoint `json:"egress"`
		Listener          proxyRuntimeListener `json:"listener"`
		ChainPlan         map[string]any       `json:"chain_plan"`
		ErrorMessage      string               `json:"error_message"`
		Labels            map[string]string    `json:"labels"`
	} `json:"lease"`
	Egress    proxyRuntimeEndpoint `json:"egress"`
	ChainPlan map[string]any       `json:"chain_plan"`
	Pool      struct {
		Endpoints []map[string]any `json:"endpoints"`
	} `json:"pool"`
}

type proxyRuntimeSession struct {
	SessionID  string `json:"session_id"`
	ProviderID string `json:"provider_id"`
}

type proxyRuntimeEndpoint struct {
	ID           string            `json:"id"`
	Protocol     string            `json:"protocol"`
	Host         string            `json:"host"`
	Port         int               `json:"port"`
	SessionID    string            `json:"session_id"`
	ProviderID   string            `json:"provider_id"`
	UpstreamKind string            `json:"upstream_kind"`
	RotationMode string            `json:"rotation_mode"`
	Labels       map[string]string `json:"labels"`
}

type proxyRuntimeListener struct {
	ListenerID string `json:"listener_id"`
	Kind       string `json:"kind"`
	ListenAddr string `json:"listen_addr"`
	Protocol   string `json:"protocol"`
	RouteID    string `json:"route_id"`
	Managed    bool   `json:"managed"`
}

type proxyRuntimeExitIPResponse struct {
	ProxyExitIP struct {
		IP           string `json:"ip"`
		ErrorMessage string `json:"error_message"`
	} `json:"proxy_exit_ip"`
}

type proxyRuntimeGeoResponse struct {
	ProxyExitGeo struct {
		IP           string `json:"ip"`
		CountryCode  string `json:"country_code"`
		Region       string `json:"region"`
		City         string `json:"city"`
		ErrorMessage string `json:"error_message"`
	} `json:"proxy_exit_geo"`
}

type proxyRuntimeFraudResponse struct {
	Check struct {
		IP             string   `json:"ip"`
		NetworkKind    string   `json:"network_kind"`
		AnonymizerKind string   `json:"anonymizer_kind"`
		RiskLevel      string   `json:"risk_level"`
		RiskScore      float64  `json:"risk_score"`
		RiskSignals    []string `json:"risk_signals"`
		CountryCode    string   `json:"country_code"`
		Region         string   `json:"region"`
		City           string   `json:"city"`
		ErrorMessage   string   `json:"error_message"`
	} `json:"check"`
}

type proxyRuntimeConnectivityResponse struct {
	Check struct {
		TargetURL    string `json:"target_url"`
		Host         string `json:"host"`
		Reachable    bool   `json:"reachable"`
		StatusCode   int    `json:"status_code"`
		LatencyMS    int    `json:"latency_ms"`
		ErrorMessage string `json:"error_message"`
	} `json:"check"`
}
