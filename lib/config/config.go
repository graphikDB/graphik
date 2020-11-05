package config

type Config struct {
	Port int      `json:"port"`
	JWKs []string `json:"jwks"`
	Cors *Cors    `json:"cors"`
	Raft *Raft    `json:"raft"`
}

type Raft struct {
	DBPath string `json:"dbPath"`
	Bind   string `json:"bind"`
	Join   string `json:"join"`
	NodeID string `json:"nodeID"`
}

type Cors struct {
	// Normalized list of plain allowed origins
	AllowedOrigins []string
	// Normalized list of allowed headers
	AllowedHeaders []string
	// Normalized list of allowed methods
	AllowedMethods []string
}
