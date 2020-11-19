package main

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/flags"
	"github.com/autom8ter/graphik/graph"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/machine"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

func main() {
	run(context.Background(), flags.Global)
}

func run(ctx context.Context, cfg *flags.Flags) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	m := machine.New(ctx)
	defer m.Close()
	g, err := graph.NewGraphStore(ctx, cfg)
	if err != nil {
		logger.Error("failed to create graph", zap.Error(err))
		return
	}
	defer g.Close()
	router := http.NewServeMux()
	if cfg.Metrics {
		router.Handle("/metrics", promhttp.Handler())
		router.HandleFunc("/debug/pprof/", pprof.Index)
		router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		router.HandleFunc("/debug/pprof/profile", pprof.Profile)
		router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	server := &http.Server{
		Handler: router,
	}

	m.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.BindHTTP)
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
			grpc_validator.UnaryServerInterceptor(),
			g.Unary(),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger.Logger()),
			grpc_validator.StreamServerInterceptor(),
			g.Stream(),
			grpc_recovery.StreamServerInterceptor(),
		),
	)

	apipb.RegisterGraphServiceServer(gserver, g)
	grpc_prometheus.Register(gserver)

	m.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.BindGrpc)
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
		m.Cancel()
		break
	case <-ctx.Done():
		m.Cancel()
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
	m.Wait()
	logger.Info("shutdown successful")
}
