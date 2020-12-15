package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/autom8ter/machine"
	"github.com/graphikDB/graphik/database"
	"github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/gql"
	"github.com/graphikDB/graphik/helpers"
	"github.com/graphikDB/graphik/logger"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/soheilhy/cmux"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var global = &apipb.Flags{}

func init() {
	godotenv.Load()
	pflag.CommandLine.StringVar(&global.StoragePath, "storage", helpers.EnvOr("GRAPHIK_STORAGE_PATH", "/tmp/graphik"), "persistant storage path (env: GRAPHIK_STORAGE_PATH)")
	pflag.CommandLine.StringVar(&global.OpenIdDiscovery, "open-id", helpers.EnvOr("GRAPHIK_OPEN_ID", ""), "open id connect discovery uri ex: https://accounts.google.com/.well-known/openid-configuration (env: GRAPHIK_OPEN_ID)")
	pflag.CommandLine.BoolVar(&global.Metrics, "metrics", helpers.BoolEnvOr("GRAPHIK_METRICS", true), "enable prometheus & pprof metrics (emv: GRAPHIK_METRICS = true)")
	pflag.CommandLine.StringSliceVar(&global.AllowHeaders, "allow-headers", helpers.StringSliceEnvOr("GRAPHIK_ALLOW_HEADERS", []string{"*"}), "cors allow headers (env: GRAPHIK_ALLOW_HEADERS)")
	pflag.CommandLine.StringSliceVar(&global.AllowOrigins, "allow-origins", helpers.StringSliceEnvOr("GRAPHIK_ALLOW_ORIGINS", []string{"*"}), "cors allow origins (env: GRAPHIK_ALLOW_ORIGINS)")
	pflag.CommandLine.StringSliceVar(&global.AllowMethods, "allow-methods", helpers.StringSliceEnvOr("GRAPHIK_ALLOW_METHODS", []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"}), "cors allow methods (env: GRAPHIK_ALLOW_METHODS)")
	pflag.CommandLine.StringSliceVar(&global.RootUsers, "root-users", helpers.StringSliceEnvOr("GRAPHIK_ROOT_USERS", nil), "a list of email addresses that bypass registered authorizers (env: GRAPHIK_ROOT_USERS)")
	pflag.CommandLine.StringVar(&global.TlsCert, "tls-cert", helpers.EnvOr("GRAPHIK_TLS_CERT", ""), "path to tls certificate (env: GRAPHIK_TLS_CERT)")
	pflag.CommandLine.StringVar(&global.TlsKey, "tls-key", helpers.EnvOr("GRAPHIK_TLS_KEY", ""), "path to tls key (env: GRAPHIK_TLS_KEY)")
	pflag.CommandLine.BoolVar(&global.RequireRequestAuthorizers, "require-request-authorizers", helpers.BoolEnvOr("GRAPHIK_REQUIRE_REQUEST_AUTHORIZERS", false), "require request authorizers for all methods/endpoints (env: GRAPHIK_REQUIRE_REQUEST_AUTHORIZERS)")
	pflag.CommandLine.BoolVar(&global.RequireResponseAuthorizers, "require-response-authorizers", helpers.BoolEnvOr("GRAPHIK_REQUIRE_RESPONSE_AUTHORIZERS", false), "require request authorizers for all methods/endpoints (env: GRAPHIK_REQUIRE_RESPONSE_AUTHORIZERS)")
	pflag.CommandLine.StringVar(&global.PlaygroundClientId, "playground-client-id", helpers.EnvOr("GRAPHIK_PLAYGROUND_CLIENT_ID", ""), "playground oauth client id (env: GRAPHIK_PLAYGROUND_CLIENT_ID)")
	pflag.CommandLine.StringVar(&global.PlaygroundClientSecret, "playground-client-secret", helpers.EnvOr("GRAPHIK_PLAYGROUND_CLIENT_SECRET", ""), "playground oauth client secret (env: GRAPHIK_PLAYGROUND_CLIENT_SECRET)")
	pflag.CommandLine.StringVar(&global.PlaygroundRedirect, "playground-redirect", helpers.EnvOr("GRAPHIK_PLAYGROUND_REDIRECT", ""), "playground oauth redirect (env: GRAPHIK_PLAYGROUND_REDIRECT)")
	pflag.Parse()
}

func main() {
	run(context.Background(), global)
}

const bind = ":7820"
const metricsBind = ":7821"

func run(ctx context.Context, cfg *apipb.Flags) {
	if cfg.OpenIdDiscovery == "" {
		logger.Error("empty open-id connect discovery --open-id", zap.String("usage", pflag.CommandLine.Lookup("open-id").Usage))
		return
	}
	if len(cfg.GetRootUsers()) == 0 {
		logger.Error("zero root users", zap.String("usage", pflag.CommandLine.Lookup("root-users").Usage))
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	m := machine.New(ctx)
	g, err := database.NewGraph(ctx, cfg)
	if err != nil {
		logger.Error("failed to create graph", zap.Error(err))
		return
	}
	defer g.Close()
	var lis net.Listener
	if global.TlsCert != "" && global.TlsKey != "" {
		cer, err := tls.LoadX509KeyPair(global.TlsCert, global.TlsKey)
		if err != nil {
			log.Println(err)
			return
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		lis, err = tls.Listen("tcp", bind, config)
		if err != nil {
			logger.Error("failed to create tls server listener", zap.Error(err))
			return
		}
	} else {
		lis, err = net.Listen("tcp", bind)
		if err != nil {
			logger.Error("failed to create server listener", zap.Error(err))
			return
		}
	}
	defer lis.Close()
	self := fmt.Sprintf("localhost%v", bind)
	conn, err := grpc.DialContext(ctx, self, grpc.WithInsecure())
	if err != nil {
		logger.Error("failed to setup graphql endpoint", zap.Error(err))
		return
	}
	var config *oauth2.Config
	if global.PlaygroundClientId != "" {
		resp, err := http.DefaultClient.Get(global.OpenIdDiscovery)
		if err != nil {
			logger.Error("failed to get oidc", zap.Error(err))
			return
		}
		defer resp.Body.Close()
		var openID = map[string]interface{}{}
		bits, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Error("failed to get oidc", zap.Error(err))
			return
		}
		if err := json.Unmarshal(bits, &openID); err != nil {
			logger.Error("failed to get oidc", zap.Error(err))
			return
		}
		config = &oauth2.Config{
			ClientID:     global.PlaygroundClientId,
			ClientSecret: global.PlaygroundClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  openID["authorization_endpoint"].(string),
				TokenURL: openID["token_endpoint"].(string),
			},
			RedirectURL: global.PlaygroundRedirect,
			Scopes:      []string{"openid", "email", "profile"},
		}
	}
	resolver := gql.NewResolver(m, apipb.NewDatabaseServiceClient(conn), cors.New(cors.Options{
		AllowedOrigins: global.AllowOrigins,
		AllowedMethods: global.AllowMethods,
		AllowedHeaders: global.AllowHeaders,
	}), config)
	mux := http.NewServeMux()
	mux.Handle("/", resolver.QueryHandler())
	if config != nil {
		mux.Handle("/playground", resolver.Playground())
		mux.Handle("/playground/callback", resolver.PlaygroundCallback("/playground"))
	}
	httpServer := &http.Server{
		Handler: mux,
	}
	cm := cmux.New(lis)
	m.Go(func(routine machine.Routine) {
		hmatcher := cm.Match(cmux.HTTP1())
		defer hmatcher.Close()
		logger.Info("starting http server",
			zap.String("address", hmatcher.Addr().String()),
		)
		if err := httpServer.Serve(hmatcher); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failure", zap.Error(err))
		}
	})
	var metricServer *http.Server
	if cfg.Metrics {
		router := http.NewServeMux()
		router.Handle("/metrics", promhttp.Handler())
		router.HandleFunc("/debug/pprof/", pprof.Index)
		router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		router.HandleFunc("/debug/pprof/profile", pprof.Profile)
		router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		router.HandleFunc("/debug/pprof/trace", pprof.Trace)
		metricServer = &http.Server{Handler: router}
		m.Go(func(routine machine.Routine) {
			metricLis, err := net.Listen("tcp", metricsBind)
			if err != nil {
				logger.Error("metrics server failure", zap.Error(err))
				return
			}
			defer metricLis.Close()
			logger.Info("starting metrics server", zap.String("address", metricLis.Addr().String()))
			if err := metricServer.Serve(metricLis); err != nil && err != http.ErrServerClosed {
				logger.Error("metrics server failure", zap.Error(err))
			}
		})
	}
	gserver := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger.Logger()),
			grpc_validator.UnaryServerInterceptor(),
			g.UnaryInterceptor(),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger.Logger()),
			grpc_validator.StreamServerInterceptor(),
			g.StreamInterceptor(),
			grpc_recovery.StreamServerInterceptor(),
		),
	)
	apipb.RegisterDatabaseServiceServer(gserver, g)
	reflection.Register(gserver)
	grpc_prometheus.Register(gserver)
	m.Go(func(routine machine.Routine) {
		gmatcher := cm.Match(cmux.HTTP2())
		defer gmatcher.Close()
		logger.Info("starting gRPC server", zap.String("address", lis.Addr().String()))
		if err := gserver.Serve(gmatcher); err != nil && err != http.ErrServerClosed && !strings.Contains(err.Error(), "mux: listener closed") {
			logger.Error("gRPC server failure", zap.Error(err))
		}
	})
	m.Go(func(routine machine.Routine) {
		if err := cm.Serve(); err != nil && !strings.Contains(err.Error(), "closed network connection") {
			logger.Error("listener mux error", zap.Error(err))
		}
	})
	select {
	case <-interrupt:
		m.Cancel()
		break
	case <-ctx.Done():
		m.Cancel()
		break
	}

	logger.Warn("shutdown signal received")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	_ = httpServer.Shutdown(shutdownCtx)
	if metricServer != nil {
		_ = metricServer.Shutdown(shutdownCtx)
	}
	stopped := make(chan struct{})
	go func() {
		gserver.GracefulStop()
		close(stopped)
	}()

	t := time.NewTimer(10 * time.Second)
	select {
	case <-t.C:
		gserver.Stop()
	case <-stopped:
		t.Stop()
	}
	m.Wait()
	logger.Info("shutdown successful")
}
