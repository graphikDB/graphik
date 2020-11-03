package config

type Config struct {
	Port    int    `json:"port"`
	JWKsURL string `json:"jwks_url"`
	Raft    *Raft  `json:"raft"`
}

type Raft struct {
	DBPath string `json:"dbPath"`
	Bind   int    `json:"bind"`
	Join   string `json:"join"`
	NodeID string `json:"nodeID"`
}
