package main

import (
	"context"
	"fmt"
	"github.com/autom8ter/graphik/database"
	"github.com/autom8ter/graphik/gen/go"
	"github.com/autom8ter/graphik/gql"
	"github.com/autom8ter/graphik/helpers"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/machine"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	pflag.CommandLine.BoolVar(&global.Metrics, "metrics", boolEnvOr("GRAPHIK_METRICS", true), "enable prometheus & pprof metrics (emv: GRAPHIK_METRICS = true)")
	pflag.CommandLine.StringSliceVar(&global.Authorizers, "authorizers", stringSliceEnvOr("GRAPHIK_AUTHORIZERS", nil), "registered CEL authorizers (env: GRAPHIK_AUTHORIZERS)")
	pflag.CommandLine.StringSliceVar(&global.AllowHeaders, "allow-headers", stringSliceEnvOr("GRAPHIK_ALLOW_HEADERS", []string{"*"}), "cors allow headers (env: GRAPHIK_ALLOW_HEADERS)")
	pflag.CommandLine.StringSliceVar(&global.AllowOrigins, "allow-origins", stringSliceEnvOr("GRAPHIK_ALLOW_ORIGINS", []string{"*"}), "cors allow origins (env: GRAPHIK_ALLOW_ORIGINS)")
	pflag.CommandLine.StringSliceVar(&global.AllowMethods, "allow-methods", stringSliceEnvOr("GRAPHIK_ALLOW_METHODS", []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"}), "cors allow methods (env: GRAPHIK_ALLOW_METHODS)")
	pflag.Parse()
}

func stringSliceEnvOr(key string, defaul []string) []string {
	if value := os.Getenv(key); value != "" {
		values := strings.Split(value, ",")
		if len(value) > 0 && values[0] != "" {
			return values
		}
	} else {
		return defaul
	}
	return nil
}

func boolEnvOr(key string, defaul bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "y", "t", "yes":
			return true
		default:
			return false
		}
	}
	return defaul
}

func main() {
	run(context.Background(), global)
}

const bind = ":7820"
const metricsBind = ":7821"

func run(ctx context.Context, cfg *apipb.Flags) {
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
	lis, err := net.Listen("tcp", bind)
	if err != nil {
		logger.Error("failed to create http server listener", zap.Error(err))
		return
	}
	defer lis.Close()
	self := fmt.Sprintf("localhost%v", bind)
	conn, err := grpc.DialContext(ctx, self, grpc.WithInsecure())
	if err != nil {
		logger.Error("failed to setup graphql endpoint", zap.Error(err))
		return
	}
	resolver := gql.NewResolver(ctx, apipb.NewDatabaseServiceClient(conn), cors.New(cors.Options{
		AllowedOrigins: global.AllowOrigins,
		AllowedMethods: global.AllowMethods,
		AllowedHeaders: global.AllowHeaders,
	}))
	httpServer := &http.Server{
		Handler: resolver.QueryHandler(),
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

	t := time.NewTimer(5 * time.Second)
	select {
	case <-t.C:
		gserver.Stop()
	case <-stopped:
		t.Stop()
	}
	m.Wait()
	logger.Info("shutdown successful")
}
