package main

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/autom8ter/graphik/graph"
	"github.com/autom8ter/graphik/graph/generated"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/machine"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
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
	viper.SetConfigFile("graphik.yaml")
	viper.SetDefault("server.port", 8080)
	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func main() {
	port := viper.GetInt("server.port")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	mach := machine.New(ctx)
	mux := http.NewServeMux()

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

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
			log.Printf("failed to create server listener: %s", err)
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
	logger.Info("shutdown signal received")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	_ = server.Shutdown(shutdownCtx)
	logger.Info("shutdown successful")
	mach.Wait()
}
