package main

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/interceptors"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/runtime"
	"github.com/autom8ter/graphik/service/private"
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
	"syscall"
	"time"
)

func init() {
	godotenv.Load()
	pflag.CommandLine.StringVar(&cfg.Grpc.Bind, "grpc.bind", ":7820", "")
	pflag.CommandLine.StringVar(&cfg.Http.Bind, "http.bind", ":7830", "")
	pflag.CommandLine.StringSliceVar(&cfg.Http.AllowedHeaders, "http.headers", nil, "cors allowed headers")
	pflag.CommandLine.StringSliceVar(&cfg.Http.AllowedMethods, "http.methods", nil, "cors allowed methods")
	pflag.CommandLine.StringSliceVar(&cfg.Http.AllowedOrigins, "http.origins", nil, "cors allowed origins")
	pflag.CommandLine.StringVar(&cfg.Raft.Bind, "raft.bind", "localhost:7840", "")
	pflag.CommandLine.StringVar(&cfg.Raft.NodeId, "raft.nodeid", "default", "")
	pflag.CommandLine.StringVar(&cfg.Raft.StoragePath, "raft.storage.path", "/tmp/graphik", "")
	pflag.CommandLine.StringSliceVar(&cfg.Auth.JwksSources, "auth.jwks", nil, "authorizaed jwks uris")
	pflag.CommandLine.StringSliceVar(&cfg.Auth.AuthExpressions, "auth.expressions", nil, "auth middleware expressions")
}

var (
	cfg = &apipb.Config{
		Http: &apipb.HTTPConfig{},
		Grpc: &apipb.GRPCConfig{},
		Raft: &apipb.RaftConfig{},
		Auth: &apipb.AuthConfig{},
	}
)

func main() {
	pflag.Parse()
	cfg.SetDefaults()
	run(context.Background(), cfg)
}

func run(ctx context.Context, config *apipb.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	runtim, err := runtime.New(ctx, config)
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
		lis, err := net.Listen("tcp", config.GetHttp().GetBind())
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

	privateService := private.NewService(runtim)

	apipb.RegisterPrivateServiceServer(gserver, privateService)
	grpc_prometheus.Register(gserver)

	runtim.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", config.GetGrpc().GetBind())
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
