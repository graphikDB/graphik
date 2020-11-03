package config

type Config struct {
	Port    int    `json:"port"`
	JWKs []string `json:"jwks"`
	Raft    *Raft  `json:"raft"`
}

type Raft struct {
	DBPath string `json:"dbPath"`
	Bind   int    `json:"bind"`
	Join   string `json:"join"`
	NodeID string `json:"nodeID"`
}
