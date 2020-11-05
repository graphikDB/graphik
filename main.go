package main

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/config"
	"github.com/autom8ter/graphik/jwks"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/runtime"
	"github.com/autom8ter/graphik/service/private"
	"github.com/autom8ter/machine"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
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

const version = "0.0.0"

func init() {
	pflag.CommandLine.StringVar(&cfg.GrpcBind, "grpc.bind", ":7686", "bind local port to gRPC server")
	pflag.CommandLine.StringVar(&cfg.HTTPBind, "http.bind", ":7687", "bind local port to http server")
	pflag.CommandLine.StringVar(&cfg.Raft.DBPath, "raft.path", "/tmp/graphik", "path to database folder")
	pflag.CommandLine.StringVar(&cfg.Raft.Bind, "raft.bind", "localhost:8090", "bind raft protocol to local port")
	pflag.CommandLine.StringVar(&cfg.Raft.Join, "raft.join", "", "join raft cluster leader")
	pflag.CommandLine.StringVar(&cfg.Raft.NodeID, "raft.id", "main", "unique raft node id")
	pflag.CommandLine.StringToStringVar(&cfg.JWKs, "jwks", nil, "remote json web key set(s)")
}

var (
	cfg = &config.Config{
		Raft: &config.Raft{},
		Cors: &config.Cors{},
	}
)

func main() {
	pflag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	mach := machine.New(ctx)
	j, err := jwks.New(cfg.JWKs)
	if err != nil {
		logger.Error("failed to fetch jwks", zap.Error(err))
		return
	}
	runt, err := runtime.New(
		runtime.WithLeader(cfg.Raft.Join == ""),
		runtime.WithID(cfg.Raft.NodeID),
		runtime.WithBindAddr(cfg.Raft.Bind),
		runtime.WithRaftDir(cfg.Raft.DBPath),
		runtime.WithJWKS(j),
		runtime.WithMachine(mach),
	)
	if err != nil {
		logger.Error("failed to create raft store", zap.Error(err))
		return
	}
	if cfg.Raft.Join != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		conn, err := grpc.DialContext(ctx, cfg.Raft.Join, grpc.WithInsecure())
		if err != nil {
			logger.Error("failed to join raft cluster", zap.Error(err))
			return
		}
		client := apipb.NewPrivateServiceClient(conn)
		_, err = client.JoinCluster(ctx, &apipb.JoinClusterRequest{
			NodeId:  cfg.Raft.NodeID,
			Address: cfg.Raft.Bind,
		})
		if err != nil {
			logger.Error("failed to join cluster", zap.Error(err))
			return
		}
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

	mach.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.HTTPBind)
		if err != nil {
			logger.Error("failed to create http server listener", zap.Error(err))
			return
		}
		defer lis.Close()
		logger.Info("starting http server",
			zap.String("address", lis.Addr().String()),
			zap.String("version", version),
		)
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failure", zap.Error(err))
		}
	})

	gserver := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger.Logger()),
			grpc_auth.UnaryServerInterceptor(runt.AuthMiddleware()),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger.Logger()),
			grpc_auth.StreamServerInterceptor(runt.AuthMiddleware()),
			grpc_recovery.StreamServerInterceptor(),
		),
	)
	privateService := private.NewService(runt)
	apipb.RegisterPrivateServiceServer(gserver, privateService)
	grpc_prometheus.Register(gserver)

	mach.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.GrpcBind)
		if err != nil {
			logger.Error("failed to create gRPC server listener", zap.Error(err))
			return
		}
		defer lis.Close()
		logger.Info("starting gRPC server",
			zap.String("address", lis.Addr().String()),
			zap.String("version", version),
		)
		if err := gserver.Serve(lis); err != nil && err != http.ErrServerClosed {
			logger.Error("gRPC server failure", zap.Error(err))
		}
	})
	select {
	case <-interrupt:
		mach.Cancel()
		break
	case <-ctx.Done():
		mach.Cancel()
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
	_ = runt.Close()
	logger.Info("shutdown successful")
	mach.Wait()
}
