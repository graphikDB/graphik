package main

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/autom8ter/graphik/graph"
	"github.com/autom8ter/graphik/graph/generated"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/store"
	"github.com/autom8ter/machine"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
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
	pflag.CommandLine.StringVar(&dbPath, "path", "/tmp/graphik", "path to database folder")
	pflag.CommandLine.IntVar(&port, "port", 8080, "port to serve on")
	pflag.CommandLine.IntVar(&bind, "bind", 8081, "bind raft protocol to local port")
	pflag.CommandLine.StringVarP(&join, "join", "j", "", "join raft cluster leader")
}

var (
	port   int
	bind   int
	dbPath string
	join   string
)

func main() {
	pflag.Parse()
	port := port
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	mach := machine.New(ctx)
	mux := http.NewServeMux()
	stor, err := store.New(
		store.WithLeader(join == ""),
		store.WithBindAddr(fmt.Sprintf("localhost:%v", bind)),
		store.WithRaftDir(dbPath),
	)
	if err != nil {
		logger.Error("failed to create raft store", zap.Error(err))
		return
	}
	resolver := graph.NewResolver(mach, stor)
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	server := &http.Server{
		Handler: mux,
	}
	mach.Go(func(routine machine.Routine) {

		lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
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
	_ = resolver.Close()
	logger.Info("shutdown successful")
	mach.Wait()
}
