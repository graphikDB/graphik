package config

type Config struct {
	Port int   `json:"port"`
	Raft *Raft `json:"raft"`
	Auth *Auth `json:"auth"`
}

type Raft struct {
	DBPath string `json:"dbPath"`
	Bind   int    `json:"bind"`
	Join   string `json:"join"`
	NodeID string `json:"nodeID"`
}

type Auth struct {
	// ClientID is the application's ID.
	ClientID string `json:"client_id" yaml:"client_id"`
	// ClientSecret is the applications client secret
	ClientSecret string `json:"client_secret" yaml:"client_secret"`
	DiscoveryUrl string `json:"discovery_url" yaml:"discovery_url"`
	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURL string `json:"redirect_url" yaml:"redirect_url"`
	// Scope specifies optional requested permissions.
	Scopes []string `json:"scopes" yaml:"scopes"`
	SessionSecret string `json:"session_secret"`
}
