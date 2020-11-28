package flags

type PluginFlags struct {
	BindGrpc string
	BindHTTP string
	Metrics  bool
}

type Flags struct {
	JWKS           string
	StoragePath    string
	Metrics        bool
	Authorizers    []string
	AllowedHeaders []string
	AllowedMethods []string
	AllowedOrigins []string
}
