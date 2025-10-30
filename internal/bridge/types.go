package bridge

// Provider types
type Provider string

const (
	ProviderVercel     Provider = "vercel"
	ProviderCloudflare Provider = "cloudflare"
	ProviderRender     Provider = "render"
	ProviderNetlify    Provider = "netlify"
)

// Error codes
type ErrorCode string

const (
	ErrAuthFailed    ErrorCode = "AUTH_FAILED"
	ErrAuthRequired  ErrorCode = "AUTH_REQUIRED"
	ErrProviderError ErrorCode = "PROVIDER_ERROR"
	ErrNetworkError  ErrorCode = "NETWORK_ERROR"
	ErrInvalidParams ErrorCode = "INVALID_PARAMS"
	ErrNotFound      ErrorCode = "NOT_FOUND"
	ErrRateLimited   ErrorCode = "RATE_LIMITED"
	ErrUnsupported   ErrorCode = "UNSUPPORTED"
	ErrTimeout       ErrorCode = "TIMEOUT"
	ErrUnknown       ErrorCode = "UNKNOWN"
)

// BridgeError represents an error from the adapter
type BridgeError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Recoverable bool                   `json:"recoverable"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

func (e *BridgeError) Error() string {
	return e.Message
}

// Response is the generic adapter response
type Response struct {
	OK             bool                   `json:"ok"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Error          *BridgeError           `json:"error,omitempty"`
	AdapterVersion string                 `json:"adapter_version"`
}

// Auth types
type AuthStartParams struct {
	Provider    Provider `json:"provider"`
	CallbackURL string   `json:"callback_url,omitempty"`
}

type AuthStartData struct {
	AuthURL   string `json:"auth_url,omitempty"`
	Token     string `json:"token,omitempty"`
	ExpiresAt *int64 `json:"expires_at,omitempty"`
}

type AuthRefreshParams struct {
	Provider     Provider `json:"provider"`
	RefreshToken string   `json:"refresh_token"`
}

type AuthRefreshData struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// Config types
type FetchConfigParams struct {
	Provider  Provider `json:"provider"`
	Token     string   `json:"token"`
	ProjectID string   `json:"project_id,omitempty"`
}

type EnvVar struct {
	Key    string   `json:"key"`
	Value  string   `json:"value"`
	Target []string `json:"target"`
}

type Project struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	Framework string `json:"framework,omitempty"`
}

type BuildConfig struct {
	Command        string `json:"command"`
	OutputDir      string `json:"output_dir"`
	InstallCommand string `json:"install_command,omitempty"`
}

type FetchConfigData struct {
	Project Project     `json:"project"`
	Build   BuildConfig `json:"build"`
	Env     []EnvVar    `json:"env"`
}

// Sync types
type SyncEnvParams struct {
	Provider  Provider `json:"provider"`
	Token     string   `json:"token"`
	ProjectID string   `json:"project_id"`
	EnvVars   []EnvVar `json:"env_vars"`
}

type SyncEnvData struct {
	Synced int      `json:"synced"`
	Failed []string `json:"failed"`
}

// Deploy types
type DeployPreviewParams struct {
	Provider  Provider          `json:"provider"`
	Token     string            `json:"token"`
	ProjectID string            `json:"project_id"`
	Branch    string            `json:"branch,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
}

type DeployPreviewData struct {
	DeploymentID string `json:"deployment_id"`
	URL          string `json:"url"`
	Status       string `json:"status"`
	BuildTime    *int   `json:"build_time,omitempty"`
}

// DNS types
type DnsUpdateParams struct {
	Provider    Provider `json:"provider"`
	Token       string   `json:"token"`
	Domain      string   `json:"domain"`
	RecordType  string   `json:"record_type"`
	RecordName  string   `json:"record_name"`
	RecordValue string   `json:"record_value"`
	TTL         int      `json:"ttl,omitempty"`
}

type DnsUpdateData struct {
	RecordID        string  `json:"record_id"`
	PreviousValue   *string `json:"previous_value,omitempty"`
	PropagationTime int     `json:"propagation_time"`
}

type DnsRollbackParams struct {
	Provider   Provider `json:"provider"`
	Token      string   `json:"token"`
	RecordID   string   `json:"record_id"`
	RollbackTo string   `json:"rollback_to"`
}

type DnsRollbackData struct {
	Restored     bool   `json:"restored"`
	CurrentValue string `json:"current_value"`
}

// Capabilities types
type CapabilitiesData struct {
	AdapterName    string   `json:"adapter_name"`
	AdapterVersion string   `json:"adapter_version"`
	SupportedVerbs []string `json:"supported_verbs"`
	AuthType       string   `json:"auth_type"`
	Features       Features `json:"features"`
}

type Features struct {
	DNSManagement      bool `json:"dns_management"`
	PreviewDeployments bool `json:"preview_deployments"`
	EnvVariables       bool `json:"env_variables"`
	BuildLogs          bool `json:"build_logs"`
}
