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
	JWKS        []string
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
	pflag.CommandLine.StringVar(&Global.StoragePath, "storage", helpers.EnvOr("GRAPHIK_STORAGE_PATH", "/tmp/graphik"), "persistant storage path (env: GRAPHIK_STORAGE_PATH)")
	pflag.CommandLine.StringSliceVar(&Global.JWKS, "jwks", strings.Split(os.Getenv("GRAPHIK_JWKS_URIS"), ","), "authorized jwks uris ex: https://www.googleapis.com/oauth2/v3/certs (env: GRAPHIK_JWKS_URIS)")
	pflag.CommandLine.BoolVar(&Global.Metrics, "metrics", os.Getenv("GRAPHIK_METRICS") == "true", "enable prometheus & pprof metrics")
	pflag.CommandLine.StringSliceVar(&Global.Triggers, "triggers", strings.Split(os.Getenv("GRAPHIK_TRIGGERS"), ","), "registered triggers (env: GRAPHIK_TRIGGERS)")
	pflag.CommandLine.StringSliceVar(&Global.Authorizers, "authorizers", strings.Split(os.Getenv("GRAPHIK_AUTHORIZERS"), ","), "registered authorizers (env: GRAPHIK_AUTHORIZERS)")
	pflag.Parse()
}
