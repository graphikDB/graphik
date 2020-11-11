package main

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/interceptors"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/runtime"
	"github.com/autom8ter/graphik/service"
	"github.com/autom8ter/machine"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	godotenv.Load()
	cfg := &apipb.Config{
		Http: &apipb.HTTPConfig{},
		Grpc: &apipb.GRPCConfig{},
		Raft: &apipb.RaftConfig{},
		Auth: &apipb.AuthConfig{},
	}
	pflag.CommandLine.StringVar(&cfg.Grpc.Bind, "grpc.bind", ":7820", "grpc server bind address")
	pflag.CommandLine.StringVar(&cfg.Http.Bind, "http.bind", ":7830", "http server bind address")
	pflag.CommandLine.StringSliceVar(&cfg.Http.AllowedHeaders, "http.headers", strings.Split(os.Getenv("GRAPHIK_HTTP_HEADERS"), ","), "cors allowed headers (env: GRAPHIK_HTTP_HEADERS)")
	pflag.CommandLine.StringSliceVar(&cfg.Http.AllowedMethods, "http.methods", strings.Split(os.Getenv("GRAPHIK_HTTP_METHODS"), ","), "cors allowed methods (env: GRAPHIK_HTTP_METHODS)")
	pflag.CommandLine.StringSliceVar(&cfg.Http.AllowedOrigins, "http.origins", strings.Split(os.Getenv("GRAPHIK_HTTP_ORIGINS"), ","), "cors allowed origins (env: GRAPHIK_HTTP_ORIGINS)")
	pflag.CommandLine.StringVar(&cfg.Raft.Bind, "raft.bind", "localhost:7840", "raft protocol bind address")
	pflag.CommandLine.StringVar(&cfg.Raft.NodeId, "raft.nodeid", os.Getenv("GRAPHIK_RAFT_ID"), "raft node id (env: GRAPHIK_RAFT_ID)")
	pflag.CommandLine.StringVar(&cfg.Raft.StoragePath, "raft.storage.path", "/tmp/graphik", "raft storage path")
	pflag.CommandLine.StringSliceVar(&cfg.Auth.JwksSources, "auth.jwks", strings.Split(os.Getenv("GRAPHIK_JWKS_URIS"), ","), "authorizaed jwks uris ex: https://www.googleapis.com/oauth2/v3/certs (env: GRAPHIK_JWKS_URIS)")
	pflag.CommandLine.StringSliceVar(&cfg.Auth.AuthExpressions, "auth.expressions", strings.Split(os.Getenv("GRAPHIK_AUTH_EXPRESSIONS"), ","), "auth middleware expressions (env: GRAPHIK_AUTH_EXPRESSIONS)")

	pflag.Parse()
	cfg.SetDefaults()
	run(context.Background(), cfg)
}

func run(ctx context.Context, cfg *apipb.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	runtim, err := runtime.New(ctx, cfg)
	if err != nil {
		logger.Error("failed to create runtime", zap.Error(err))
		return
	}
	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Handler: router,
	}

	runtim.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.GetHttp().GetBind())
		if err != nil {
			logger.Error("failed to create http server listener", zap.Error(err))
			return
		}
		defer lis.Close()
		logger.Info("starting http server",
			zap.String("address", lis.Addr().String()),
		)
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failure", zap.Error(err))
		}
	})

	gserver := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger.Logger()),
			interceptors.UnaryAuth(runtim),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger.Logger()),
			interceptors.StreamAuth(runtim),
			grpc_recovery.StreamServerInterceptor(),
		),
	)

	apipb.RegisterConfigServiceServer(gserver, service.NewConfig(runtim))
	apipb.RegisterGraphServiceServer(gserver, service.NewGraph(runtim))
	grpc_prometheus.Register(gserver)

	runtim.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.GetGrpc().GetBind())
		if err != nil {
			logger.Error("failed to create gRPC server listener", zap.Error(err))
			return
		}
		defer lis.Close()
		logger.Info("starting gRPC server",
			zap.String("address", lis.Addr().String()),
		)
		if err := gserver.Serve(lis); err != nil && err != http.ErrServerClosed {
			logger.Error("gRPC server failure", zap.Error(err))
		}
	})
	select {
	case <-interrupt:
		runtim.Cancel()
		break
	case <-ctx.Done():
		runtim.Cancel()
		break
	}

	logger.Warn("shutdown signal received")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	_ = server.Shutdown(shutdownCtx)
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
	_ = runtim.Close()
	logger.Info("shutdown successful")
	runtim.Wait()
}
