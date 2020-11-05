package main

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/lib/config"
	"github.com/autom8ter/graphik/lib/jwks"
	"github.com/autom8ter/graphik/lib/logger"
	"github.com/autom8ter/graphik/lib/runtime"
	"github.com/autom8ter/machine"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
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
	pflag.CommandLine.IntVar(&cfg.Port, "port", 8080, "port to serve on")
	pflag.CommandLine.StringVar(&cfg.Raft.DBPath, "raft.path", "/tmp/graphik", "path to database folder")
	pflag.CommandLine.StringVar(&cfg.Raft.Bind, "raft.bind", ":8081", "bind raft protocol to local port")
	pflag.CommandLine.StringVar(&cfg.Raft.Join, "raft.join", "", "join raft cluster leader")
	pflag.CommandLine.StringVar(&cfg.Raft.NodeID, "raft.id", "main", "unique raft node id")
	pflag.CommandLine.StringSliceVar(&cfg.JWKs, "jwks", nil, "remote json web key set(s)")
	pflag.CommandLine.StringSliceVar(&cfg.Cors.AllowedHeaders, "cors.headers", nil, "allowed cors headers")
	pflag.CommandLine.StringSliceVar(&cfg.Cors.AllowedMethods, "cors.methods", nil, "allowed cors methods")
	pflag.CommandLine.StringSliceVar(&cfg.Cors.AllowedOrigins, "cors.origins", nil, "allowed cors origins")
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
	router := mux.NewRouter()
	j, err := jwks.New(cfg.JWKs)
	if err != nil {
		logger.Error("failed to fetch jwks", zap.Error(err))
		return
	}
	stor, err := runtime.New(
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
		client := apipb.NewAdminServiceClient(conn)
		_, err = client.RaftJoin(ctx, &apipb.RaftJoinRequest{
			NodeId:  cfg.Raft.NodeID,
			Address: cfg.Raft.Bind,
		})
		if err != nil {
			logger.Error("failed to join cluster", zap.Error(err))
			return
		}
	}
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   cfg.Cors.AllowedOrigins,
		AllowedMethods:   cfg.Cors.AllowedMethods,
		AllowedHeaders:   cfg.Cors.AllowedHeaders,
		AllowCredentials: true,
	}).Handler)

	router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
	router.HandleFunc("/debug/pprof/", pprof.Index).Methods(http.MethodGet)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline).Methods(http.MethodGet)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile).Methods(http.MethodGet)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol).Methods(http.MethodGet)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace).Methods(http.MethodGet)

	server := &http.Server{
		Handler: router,
	}
	mach.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.Port))
		if err != nil {
			logger.Error("failed to create server listener", zap.Error(err))
			return
		}
		defer lis.Close()
		logger.Info("starting graphql server",
			zap.String("address", lis.Addr().String()),
			zap.String("version", version),
		)
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			logger.Error("server failure", zap.Error(err))
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
	_ = stor.Close()
	logger.Info("shutdown successful")
	mach.Wait()
}
