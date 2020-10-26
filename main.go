package main

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/autom8ter/dagger"
	"github.com/autom8ter/graphik/graph"
	"github.com/autom8ter/graphik/graph/generated"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/machine"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
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
	viper.SetConfigFile("graphik.yaml")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.path", "/tmp/graphik/graphik.db")
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

	{
		f, err := getDBFile()
		if err != nil {
			logger.Error("failed to open database", zap.Error(err))
			return
		}
		if err := dagger.ImportJSON(f); err != nil {
			logger.Error("failed to import database", zap.Error(err))
		}
		logger.Info("node types", zap.Strings("types", dagger.NodeTypes()))
	}

	mach := machine.New(ctx)
	mux := http.NewServeMux()

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: graph.NewResolver(mach)}))

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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	_ = server.Shutdown(shutdownCtx)

	{
		f, err := getDBFile()
		if err != nil {
			logger.Error("failed to open database", zap.Error(err))
		}
		if err := f.Truncate(0); err != nil {
			logger.Error("failed to export graph", zap.Error(err))
		}
		if err := dagger.ExportJSON(f); err != nil {
			logger.Error("failed to export graph", zap.Error(err))
		}
	}
	logger.Info("shutdown successful")
	mach.Wait()
}


func getDBFile() (*os.File, error) {
	os.MkdirAll("/tmp/graphik/", os.ModePerm)
	return os.OpenFile(viper.GetString("database.path"), os.O_CREATE|os.O_RDWR, os.ModePerm)
}