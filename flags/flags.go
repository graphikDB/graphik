package flags

import (
	"github.com/autom8ter/graphik/helpers"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

type PluginFlags struct {
	BindGrpc string
	BindHTTP string
	Metrics  bool
}

type Flags struct {
	BindGrpc    string
	BindHTTP    string
	BindRaft    string
	HttpHeaders []string
	HttpMethods []string
	HttpOrigins []string
	JWKS        []string
	RaftID      string
	JoinRaft    string
	StoragePath string
	Metrics     bool
	Triggers    []string
	Authorizers []string
}

var Global = &Flags{}

func init() {
	godotenv.Load()
	pflag.CommandLine.StringVar(&Global.BindGrpc, "grpc.bind", ":7820", "grpc server bind address")
	pflag.CommandLine.StringVar(&Global.BindHTTP, "http.bind", ":7830", "http server bind address")
	pflag.CommandLine.StringVar(&Global.BindRaft, "raft.bind", "localhost:7840", "raft protocol bind address")
	pflag.CommandLine.StringVar(&Global.JoinRaft, "raft.join", "", "join raft at target address")
	pflag.CommandLine.StringSliceVar(&Global.HttpHeaders, "http.headers", strings.Split(os.Getenv("GRAPHIK_HTTP_HEADERS"), ","), "cors allowed headers (env: GRAPHIK_HTTP_HEADERS)")
	pflag.CommandLine.StringSliceVar(&Global.HttpMethods, "http.methods", strings.Split(os.Getenv("GRAPHIK_HTTP_METHODS"), ","), "cors allowed methods (env: GRAPHIK_HTTP_METHODS)")
	pflag.CommandLine.StringSliceVar(&Global.HttpOrigins, "http.origins", strings.Split(os.Getenv("GRAPHIK_HTTP_ORIGINS"), ","), "cors allowed origins (env: GRAPHIK_HTTP_ORIGINS)")
	pflag.CommandLine.StringVar(&Global.RaftID, "raft.id", helpers.EnvOr("GRAPHIK_RAFT_ID", "leader"), "raft node id (env: GRAPHIK_RAFT_ID)")
	pflag.CommandLine.StringVar(&Global.StoragePath, "storage", helpers.EnvOr("GRAPHIK_STORAGE_PATH", "/tmp/graphik"), "persistant storage path (env: GRAPHIK_STORAGE_PATH)")
	pflag.CommandLine.StringSliceVar(&Global.JWKS, "jwks", strings.Split(os.Getenv("GRAPHIK_JWKS_URIS"), ","), "authorized jwks uris ex: https://www.googleapis.com/oauth2/v3/certs (env: GRAPHIK_JWKS_URIS)")
	pflag.CommandLine.BoolVar(&Global.Metrics, "metrics", os.Getenv("GRAPHIK_METRICS") == "true", "enable prometheus & pprof metrics")
	pflag.CommandLine.StringSliceVar(&Global.Triggers, "triggers", strings.Split(os.Getenv("GRAPHIK_TRIGGERS"), ","), "registered triggers (env: GRAPHIK_TRIGGERS)")
	pflag.CommandLine.StringSliceVar(&Global.Authorizers, "authorizers", strings.Split(os.Getenv("GRAPHIK_AUTHORIZERS"), ","), "registered authorizers (env: GRAPHIK_AUTHORIZERS)")
	pflag.Parse()
}
