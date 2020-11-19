package graphik

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/flags"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/machine"
	"github.com/golang/protobuf/ptypes/empty"
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

// TriggerFunc is an optional/custom external plugin that when added to a Graphik instance, mutates objects at runtime before & after state changes.
// It should be deployed as a side car to a graphik instance(see Serve())
type TriggerFunc func(ctx context.Context, trigger *apipb.Trigger) (*apipb.StateChange, error)

func NewTrigger(trigger func(ctx context.Context, trigger *apipb.Trigger) (*apipb.StateChange, error)) TriggerFunc {
	return TriggerFunc(trigger)
}

func (t TriggerFunc) HandleTrigger(ctx context.Context, trigger *apipb.Trigger) (*apipb.StateChange, error) {
	return t(ctx, trigger)
}

func (t TriggerFunc) Ping(ctx context.Context, _ *empty.Empty) (*apipb.Pong, error) {
	return &apipb.Pong{
		Message: "PONG",
	}, nil
}

func (t TriggerFunc) Serve(ctx context.Context, cfg *flags.PluginFlags) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	mach := machine.New(ctx)
	defer mach.Close()
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

	mach.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.BindHTTP)
		if err != nil {
			logger.Error("failed to create http server listener", zap.Error(err))
			return
		}
		defer lis.Close()
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failure", zap.Error(err))
		}
	})

	gserver := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger.Logger()),
			grpc_validator.UnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger.Logger()),
			grpc_validator.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(),
		),
	)

	apipb.RegisterTriggerServiceServer(gserver, t)
	grpc_prometheus.Register(gserver)

	mach.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.BindGrpc)
		if err != nil {
			logger.Error("failed to create gRPC server listener", zap.Error(err))
			return
		}
		defer lis.Close()
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	_ = server.Shutdown(shutdownCtx)
	stopped := make(chan struct{})
	go func() {
		gserver.GracefulStop()
		close(stopped)
	}()

	timer := time.NewTimer(5 * time.Second)

	select {
	case <-timer.C:
		gserver.Stop()
	case <-stopped:
		timer.Stop()
	}
	mach.Wait()
}

// AuthorizerFunc is an optional/custom external plugin that when added to a graphik instance,, authorizes inbound graph requests.
// It should be deployed as a side car to a graphik instance(see Serve())
type AuthorizerFunc func(ctx context.Context, intercept *apipb.RequestIntercept) (*apipb.Decision, error)

func NewAuthorizer(auth func(ctx context.Context, intercept *apipb.RequestIntercept) (*apipb.Decision, error)) AuthorizerFunc {
	return AuthorizerFunc(auth)
}

func (a AuthorizerFunc) Authorize(ctx context.Context, intercept *apipb.RequestIntercept) (*apipb.Decision, error) {
	return a(ctx, intercept)
}

func (a AuthorizerFunc) Ping(ctx context.Context, _ *empty.Empty) (*apipb.Pong, error) {
	return &apipb.Pong{
		Message: "PONG",
	}, nil
}

func (a AuthorizerFunc) Serve(ctx context.Context, cfg *flags.PluginFlags) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	mach := machine.New(ctx)
	defer mach.Close()
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

	mach.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.BindHTTP)
		if err != nil {
			logger.Error("failed to create http server listener", zap.Error(err))
			return
		}
		defer lis.Close()
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failure", zap.Error(err))
		}
	})

	gserver := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger.Logger()),
			grpc_validator.UnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger.Logger()),
			grpc_validator.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(),
		),
	)

	apipb.RegisterAuthorizationServiceServer(gserver, a)
	grpc_prometheus.Register(gserver)

	mach.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.BindGrpc)
		if err != nil {
			logger.Error("failed to create gRPC server listener", zap.Error(err))
			return
		}
		defer lis.Close()
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	_ = server.Shutdown(shutdownCtx)
	stopped := make(chan struct{})
	go func() {
		gserver.GracefulStop()
		close(stopped)
	}()

	timer := time.NewTimer(5 * time.Second)

	select {
	case <-timer.C:
		gserver.Stop()
	case <-stopped:
		timer.Stop()
	}
	mach.Wait()
}
