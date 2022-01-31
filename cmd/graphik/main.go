package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/autom8ter/machine"
	"github.com/graphikDB/graphik/database"
	"github.com/graphikDB/graphik/discover/k8s"
	"github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/gql"
	"github.com/graphikDB/graphik/gql/session"
	"github.com/graphikDB/graphik/helpers"
	"github.com/graphikDB/graphik/logger"
	"github.com/graphikDB/graphik/raft"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/soheilhy/cmux"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"
)

var global = &apipb.Flags{}
var uiGlobal = &apipb.UIFlags{}

func init() {
	godotenv.Load()
	pflag.CommandLine.BoolVar(&global.Debug, "debug", helpers.BoolEnvOr("GRAPHIK_DEBUG", false), "enable debug logs (env: GRAPHIK_DEBUG)")
	pflag.CommandLine.StringVar(&global.RaftSecret, "raft-secret", os.Getenv("GRAPHIK_RAFT_SECRET"), "raft cluster secret (so only authorized nodes may join cluster) (env: GRAPHIK_RAFT_SECRET)")
	pflag.CommandLine.StringVar(&global.JoinRaft, "join-raft", os.Getenv("GRAPHIK_JOIN_RAFT"), "join raft cluster at target address (env: GRAPHIK_JOIN_RAFT)")
	pflag.CommandLine.StringVar(&global.RaftPeerId, "raft-peer-id", os.Getenv("GRAPHIK_RAFT_PEER_ID"), "raft peer ID - one will be generated if not set (env: GRAPHIK_RAFT_PEER_ID)")
	pflag.CommandLine.Int64Var(&global.ListenPort, "listen-port", int64(helpers.IntEnvOr("GRAPHIK_LISTEN_PORT", 7820)), "serve gRPC & graphQL on this port (env: GRAPHIK_LISTEN_PORT)")
	pflag.CommandLine.StringVar(&global.StoragePath, "storage", helpers.EnvOr("GRAPHIK_STORAGE_PATH", "/tmp/graphik"), "persistant storage path (env: GRAPHIK_STORAGE_PATH)")
	pflag.CommandLine.StringVar(&global.OpenIdDiscovery, "open-id", helpers.EnvOr("GRAPHIK_OPEN_ID", ""), "open id connect discovery uri ex: https://accounts.google.com/.well-known/openid-configuration (env: GRAPHIK_OPEN_ID) (required)")
	pflag.CommandLine.StringSliceVar(&global.AllowHeaders, "allow-headers", helpers.StringSliceEnvOr("GRAPHIK_ALLOW_HEADERS", []string{"*"}), "cors allow headers (env: GRAPHIK_ALLOW_HEADERS)")
	pflag.CommandLine.StringSliceVar(&global.AllowOrigins, "allow-origins", helpers.StringSliceEnvOr("GRAPHIK_ALLOW_ORIGINS", []string{"*"}), "cors allow origins (env: GRAPHIK_ALLOW_ORIGINS)")
	pflag.CommandLine.StringSliceVar(&global.AllowMethods, "allow-methods", helpers.StringSliceEnvOr("GRAPHIK_ALLOW_METHODS", []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"}), "cors allow methods (env: GRAPHIK_ALLOW_METHODS)")
	pflag.CommandLine.StringSliceVar(&global.RootUsers, "root-users", helpers.StringSliceEnvOr("GRAPHIK_ROOT_USERS", nil), "a list of email addresses that bypass registered authorizers (env: GRAPHIK_ROOT_USERS)  (required)")
	pflag.CommandLine.StringVar(&global.TlsCert, "tls-cert", helpers.EnvOr("GRAPHIK_TLS_CERT", ""), "path to tls certificate (env: GRAPHIK_TLS_CERT)")
	pflag.CommandLine.StringVar(&global.TlsKey, "tls-key", helpers.EnvOr("GRAPHIK_TLS_KEY", ""), "path to tls key (env: GRAPHIK_TLS_KEY)")
	pflag.CommandLine.BoolVar(&global.RequireRequestAuthorizers, "require-request-authorizers", helpers.BoolEnvOr("GRAPHIK_REQUIRE_REQUEST_AUTHORIZERS", false), "require request authorizers for all methods/endpoints (env: GRAPHIK_REQUIRE_REQUEST_AUTHORIZERS)")
	pflag.CommandLine.BoolVar(&global.RequireResponseAuthorizers, "require-response-authorizers", helpers.BoolEnvOr("GRAPHIK_REQUIRE_RESPONSE_AUTHORIZERS", false), "require request authorizers for all methods/endpoints (env: GRAPHIK_REQUIRE_RESPONSE_AUTHORIZERS)")
	pflag.CommandLine.StringVar(&global.Environment, "environment", helpers.EnvOr("GRAPHIK_ENVIRONMENT", ""), "deployment environment (k8s) (env: GRAPHIK_ENVIRONMENT)")
	pflag.CommandLine.Int64Var(&global.RaftMaxPool, "raft-max-pool", int64(helpers.IntEnvOr("GRAPHIK_RAFT_MAX_POOL", 5)), "max nodes in pool (env: GRAPHIK_RAFT_MAX_POOL)")
	pflag.CommandLine.BoolVar(&global.MutualTls, "mutual-tls", helpers.BoolEnvOr("GRAPHIK_MUTUAL_TLS", false), "require mutual tls (env: GRAPHIK_MUTUAL_TLS)")
	pflag.CommandLine.StringVar(&global.CaCert, "ca-cert", helpers.EnvOr("GRAPHIK_CA_CERT", ""), "client CA certificate path for establishing mtls (env: GRAPHIK_CA_CERT)")

	pflag.CommandLine.BoolVar(&global.EnableUi, "enable-ui", helpers.BoolEnvOr("GRAPHIK_ENABLE_UI", true), "enable user interface (env: GRAPHIK_ENABLE_UI)")

	pflag.CommandLine.StringVar(&uiGlobal.OauthClientId, "ui-oauth-client-id", helpers.EnvOr("GRAPHIK_UI_OAUTH_CLIENT_ID", "723941275880-6i69h7d27ngmcnq02p6t8lbbgenm26um.apps.googleusercontent.com"), "user authentication: oauth client id (env: GRAPHIK_UI_OAUTH_CLIENT_ID)")
	pflag.CommandLine.StringVar(&uiGlobal.OauthClientSecret, "ui-oauth-client-secret", helpers.EnvOr("GRAPHIK_UI_OAUTH_CLIENT_SECRET", "E2ru-iJAxijisJ9RzMbloe4c"), "user authentication: oauth client secret (env: GRAPHIK_UI_OAUTH_CLIENT_SECRET)")
	pflag.CommandLine.StringVar(&uiGlobal.OauthRedirectUrl, "ui-oauth-redirect-url", helpers.EnvOr("GRAPHIK_UI_OAUTH_REDIRECT_URL", "http://localhost:7820/ui/login"), "user authentication: oauth redirect url (env: GRAPHIK_UI_OAUTH_REDIRECT_URL)")
	pflag.CommandLine.StringVar(&uiGlobal.OauthAuthorizationUrl, "ui-oauth-authorization-url", helpers.EnvOr("GRAPHIK_UI_OAUTH_AUTHORIZATION_URL", "https://accounts.google.com/o/oauth2/v2/auth"), "user authentication: oauth authorization url (env: GRAPHIK_UI_OAUTH_AUTHORIZATION_URL)")
	pflag.CommandLine.StringVar(&uiGlobal.OauthTokenUrl, "ui-oauth-token-url", helpers.EnvOr("GRAPHIK_UI_OAUTH_TOKEN_URL", "https://oauth2.googleapis.com/token"), "user authentication: token url (env: GRAPHIK_UI_OAUTH_TOKEN_URL)")
	pflag.CommandLine.StringSliceVar(&uiGlobal.OauthScopes, "ui-oauth-scopes", strings.Split(helpers.EnvOr("GRAPHIK_UI_OAUTH_SCOPES", "openid,email,profile"), ","), "user authentication: oauth scopes (env: GRAPHIK_UI_OAUTH_SCOPES)")
	pflag.CommandLine.StringVar(&uiGlobal.SessionSecret, "ui-session-secret", helpers.EnvOr("GRAPHIK_UI_SESSION_SECRET", "change-me-xxxx-xxxx"), "user authentication: session secret (env: GRAPHIK_UI_SESSION_SECRET)")

	pflag.Parse()
}

const (
	k8sEnv    = "k8s"
	leaderPod = "graphik-0"
)

func main() {
	run(context.Background(), global)
}

func run(ctx context.Context, cfg *apipb.Flags) {
	var lgger = logger.New(cfg.Debug)
	if cfg.OpenIdDiscovery == "" {
		lgger.Error("empty open-id connect discovery --open-id", zap.String("usage", pflag.CommandLine.Lookup("open-id").Usage))
		return
	}
	if len(cfg.GetRootUsers()) == 0 {
		lgger.Error("zero root users", zap.String("usage", pflag.CommandLine.Lookup("root-users").Usage))
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var (
		localRaftAddr = fmt.Sprintf("localhost:%v", global.ListenPort+1)
		adminLis      net.Listener
		m             = machine.New(ctx)
		err           error
		tlsConfig     *tls.Config
		apiLis        net.Listener
		advertise     net.Addr
	)
	defer m.Close()
	if global.TlsCert != "" && global.TlsKey != "" {
		cer, err := tls.LoadX509KeyPair(global.TlsCert, global.TlsKey)
		if err != nil {
			lgger.Error("failed to load tls config",
				zap.String("cert", global.TlsCert),
				zap.String("key", global.TlsKey),
				zap.Error(err))
			return
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cer},
		}
		if global.MutualTls {
			lgger.Debug("mtls enabled")
			lgger.Debug("loading client CA certificates")
			pemClientCA, err := ioutil.ReadFile(global.CaCert)
			if err != nil {
				lgger.Error("failed to load client CA certificate", zap.Error(err))
				return
			}

			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(pemClientCA) {
				lgger.Error("failed to add client CA certificate")
				return
			}
			tlsConfig.ClientCAs = certPool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}
	if global.Environment != "" {
		switch global.Environment {
		case k8sEnv:
			var (
				podname   = os.Getenv("POD_NAME")
				namespace = os.Getenv("POD_NAMESPACE")
				podIp     = os.Getenv("POD_IP")
			)
			if podname == "" {
				lgger.Error("expected POD_NAME environmental variable set in k8s environment")
				return
			}
			if podIp == "" {
				lgger.Error("expected POD_IP environmental variable set in k8s environment")
				return
			}
			prvider, err := k8s.NewInClusterProvider(namespace)
			if err != nil {
				lgger.Error("failed to get incluster k8s provider", zap.Error(err))
				return
			}
			pods, err := prvider.Pods(ctx)
			if err != nil {
				lgger.Error("failed to get pods", zap.Error(err))
				return
			}
			leaderIp := pods[leaderPod]
			lgger.Debug("registered k8s environment",
				zap.String("namespace", namespace),
				zap.String("podip", podIp),
				zap.String("podname", podname),
				zap.String("leaderIp", leaderIp),
				zap.Any("discovery", pods),
			)
			global.RaftPeerId = podname

			advertise := fmt.Sprintf("%s.graphik.graphik.svc.cluster.local:%v", podname, global.ListenPort+1)
			addr, err := net.ResolveTCPAddr("tcp", advertise)
			if err != nil {
				lgger.Error("failed to parse local ip address", zap.String("address", advertise), zap.Error(err))
				return
			}
			global.RaftAdvertise = addr.String()
			if podname != leaderPod {
				global.JoinRaft = fmt.Sprintf("%s:%v", leaderIp, global.ListenPort)
				localRaftAddr = addr.String()
			}
		default:
			lgger.Error("unsupported environment", zap.String("env", global.Environment))
			return
		}
	}
	{
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%v", global.ListenPort+1))
		if err != nil {
			lgger.Error("failed to create listener", zap.Error(err))
			return
		}
		adminLis, err = net.ListenTCP("tcp", addr)
		if err != nil {
			lgger.Error("failed to create listener", zap.Error(err))
			return
		}
	}
	defer adminLis.Close()
	adminMux := cmux.New(adminLis)
	hmatcher := adminMux.Match(cmux.HTTP1(), cmux.HTTP1Fast())
	defer hmatcher.Close()
	raftLis := adminMux.Match(cmux.Any())
	if tlsConfig != nil {
		raftLis = tls.NewListener(raftLis, tlsConfig)
	}
	lgger.Info("starting raft server", zap.String("address", raftLis.Addr().String()))
	defer raftLis.Close()
	var metricServer *http.Server
	{
		router := http.NewServeMux()
		router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		router.Handle("/metrics", promhttp.Handler())
		router.HandleFunc("/debug/pprof/", pprof.Index)
		router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		router.HandleFunc("/debug/pprof/profile", pprof.Profile)
		router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		router.HandleFunc("/debug/pprof/trace", pprof.Trace)
		metricServer = &http.Server{Handler: router}
	}

	m.Go(func(routine machine.Routine) {
		lgger.Info("starting metrics/admin server", zap.String("address", hmatcher.Addr().String()))
		if err := metricServer.Serve(hmatcher); err != nil && err != http.ErrServerClosed {
			lgger.Error("metrics server failure", zap.Error(err))
		}
	})
	m.Go(func(routine machine.Routine) {
		if err := adminMux.Serve(); err != nil && !strings.Contains(err.Error(), "closed network connection") {
			lgger.Error("listener mux error", zap.Error(err))
		}
	})
	lgger.Debug("creating graph")
	g, err := database.NewGraph(
		global.OpenIdDiscovery,
		database.WithStoragePath(global.StoragePath),
		database.WithRaftSecret(global.RaftSecret),
		database.WithLogger(lgger),
		database.WithMachine(m),
		database.WithRootUsers(global.RootUsers),
		database.WithRequireRequestAuthorizers(global.RequireRequestAuthorizers),
		database.WithRequireResponseAuthorizers(global.RequireResponseAuthorizers),
	)
	if err != nil {
		lgger.Error("failed to create graph", zap.Error(err))
		return
	}
	defer g.Close()
	lgger.Debug("creating api tcp listener")

	ropts := []raft.Opt{
		raft.WithIsLeader(global.JoinRaft == ""),
		raft.WithRaftDir(fmt.Sprintf("%s/raft", global.StoragePath)),
		raft.WithPeerID(global.RaftPeerId),
		raft.WithMaxPool(int(global.RaftMaxPool)),
		raft.WithRestoreSnapshotOnRestart(false),
		raft.WithTimeout(3 * time.Second),
		raft.WithDebug(global.Debug),
	}
	if global.RaftAdvertise != "" {
		advertise, err = net.ResolveTCPAddr("tcp", global.RaftAdvertise)
		if err != nil {
			lgger.Error("failed to resolve raft advertise addr", zap.Error(err))
			return
		}
		ropts = append(ropts, raft.WithAdvertiseAddr(advertise))
	}
	lgger.Debug("setting up raft")
	rft, err := raft.NewRaft(g.RaftFSM(), raftLis, ropts...)
	if err != nil {
		lgger.Error("failed to create raft", zap.Error(err))
		return
	}
	defer func() {
		if err := rft.Close(); err != nil {

		}
	}()
	lgger.Debug("successfully setup raft")
	g.SetRaft(rft)
	self := fmt.Sprintf("localhost:%v", global.ListenPort)
	conn, err := grpc.DialContext(ctx, self, grpc.WithInsecure())
	if err != nil {
		lgger.Error("failed to setup graphql endpoint", zap.Error(err))
		return
	}
	defer conn.Close()

	mux := http.NewServeMux()

	if global.EnableUi {
		if uiGlobal.OauthClientId == "" {
			lgger.Error("ui: validation error - empty oauth client id")
			return
		}
		if uiGlobal.OauthClientSecret == "" {
			lgger.Error("ui: validation error - empty oauth client secret")
			return
		}
		if uiGlobal.OauthAuthorizationUrl == "" {
			lgger.Error("ui: validation error - empty oauth authorization url")
			return
		}
		if uiGlobal.OauthTokenUrl == "" {
			lgger.Error("ui: validation error - empty oauth token url")
			return
		}
		if uiGlobal.OauthRedirectUrl == "" {
			lgger.Error("ui: validation error - empty oauth redirect url")
			return
		}
		if uiGlobal.SessionSecret == "" {
			lgger.Error("ui: validation error - empty session secret")
			return
		}
		if len(uiGlobal.OauthScopes) == 0 {
			lgger.Error("ui: validation error - empty oauth scopes")
			return
		}
		oauthConfig := &oauth2.Config{
			ClientID:     uiGlobal.OauthClientId,
			ClientSecret: uiGlobal.OauthClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  uiGlobal.OauthAuthorizationUrl,
				TokenURL: uiGlobal.OauthTokenUrl,
			},
			RedirectURL: uiGlobal.OauthRedirectUrl,
			Scopes:      uiGlobal.OauthScopes,
		}

		sessManager, err := session.GetSessionManager(oauthConfig, map[string]string{
			"name":   "cookies",
			"secret": uiGlobal.SessionSecret,
		})
		if err != nil {
			lgger.Error("ui: failed to setup session manager")
			return
		}
		resolver := gql.NewResolver(apipb.NewDatabaseServiceClient(conn), apipb.NewRaftServiceClient(conn), lgger, sessManager)
		mux.Handle("/", resolver.QueryHandler())
		mux.Handle("/ui", resolver.UIHandler())
		mux.Handle("/ui/login", resolver.OAuthCallback())
	} else {
		resolver := gql.NewResolver(apipb.NewDatabaseServiceClient(conn), apipb.NewRaftServiceClient(conn), lgger, nil)
		mux.Handle("/", resolver.QueryHandler())
	}

	httpServer := &http.Server{
		Handler: mux,
	}
	if tlsConfig != nil {
		httpServer.TLSConfig = tlsConfig
	}

	apiLis, err = net.Listen("tcp", fmt.Sprintf(":%v", global.ListenPort))
	if err != nil {
		lgger.Error("failed to create api server listener", zap.Error(err))
		return
	}
	defer apiLis.Close()
	apiMux := cmux.New(apiLis)
	m.Go(func(routine machine.Routine) {
		hmatcher := apiMux.Match(cmux.HTTP1())
		defer hmatcher.Close()
		lgger.Info("starting http server",
			zap.String("address", hmatcher.Addr().String()),
		)
		if err := httpServer.Serve(hmatcher); err != nil && err != http.ErrServerClosed {
			lgger.Error("http server failure", zap.Error(err))
		}
	})
	gopts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(lgger.Zap()),
			grpc_validator.UnaryServerInterceptor(),
			g.UnaryInterceptor(),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(lgger.Zap()),
			grpc_validator.StreamServerInterceptor(),
			g.StreamInterceptor(),
			grpc_recovery.StreamServerInterceptor(),
		),
	}
	if tlsConfig != nil {
		gopts = append(gopts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}
	gserver := grpc.NewServer(gopts...)
	apipb.RegisterDatabaseServiceServer(gserver, g)
	apipb.RegisterRaftServiceServer(gserver, g)
	reflection.Register(gserver)

	grpc_prometheus.Register(gserver)
	gmatcher := apiMux.Match(cmux.HTTP2())
	defer gmatcher.Close()
	m.Go(func(routine machine.Routine) {
		lgger.Info("starting gRPC server", zap.String("address", gmatcher.Addr().String()))
		if err := gserver.Serve(gmatcher); err != nil && err != http.ErrServerClosed && !strings.Contains(err.Error(), "mux: listener closed") {
			lgger.Error("gRPC server failure", zap.Error(err))
		}
	})
	m.Go(func(routine machine.Routine) {
		if err := apiMux.Serve(); err != nil && !strings.Contains(err.Error(), "closed network connection") {
			lgger.Error("listener mux error", zap.Error(err))
		}
	})

	m.Go(func(routine machine.Routine) {
		if global.JoinRaft != "" {
			lgger.Debug("joining raft cluster",
				zap.String("joinAddr", global.JoinRaft),
				zap.String("localAddr", localRaftAddr),
			)

			if err := join(ctx, global.JoinRaft, localRaftAddr, g, lgger); err != nil {
				lgger.Error("failed to join raft", zap.Error(err))
				cancel()
			}
		}
	})
	//select {
	//case <-interrupt:
	//	m.Close()
	//	break
	//case <-ctx.Done():
	//	m.Close()
	//	break
	//}
	//lgger.Warn("shutdown signal received")
	//shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer shutdownCancel()
	//
	//_ = httpServer.Shutdown(shutdownCtx)
	//_ = metricServer.Shutdown(shutdownCtx)
	//stopped := make(chan struct{})
	//go func() {
	//	gserver.GracefulStop()
	//	close(stopped)
	//}()
	//
	//t := time.NewTimer(10 * time.Second)
	//select {
	//case <-t.C:
	//	gserver.Stop()
	//case <-stopped:
	//	t.Stop()
	//}
	m.Wait()
	g.Close()
	lgger.Debug("shutdown successful")
}

func join(ctx context.Context, joinAddr, localAddr string, g *database.Graph, lgger *logger.Logger) error {

	leaderConn, err := grpc.DialContext(ctx, joinAddr, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to join raft")
	}
	defer leaderConn.Close()
	rclient := apipb.NewRaftServiceClient(leaderConn)
	for x := 0; x < 30; x++ {
		if ctx.Err() != nil {
			return nil
		}
		if err := pingTCP(localAddr); err != nil {
			lgger.Error("failed to join cluster - retrying", zap.Error(err),
				zap.Int("attempt", x+1),
				zap.String("joinAddr", joinAddr),
				zap.String("localAddr", localAddr),
			)
			time.Sleep(1 * time.Second)
			continue
		}
		_, err := rclient.JoinCluster(metadata.AppendToOutgoingContext(ctx, "x-graphik-raft-secret", g.RaftSecret()), &apipb.Peer{
			NodeId: g.Raft().PeerID(),
			Addr:   localAddr,
		})
		if err != nil {
			lgger.Error("failed to join cluster - retrying", zap.Error(err),
				zap.Int("attempt", x+1),
				zap.String("joinAddr", joinAddr),
				zap.String("localAddr", localAddr),
			)
			time.Sleep(1 * time.Second)
			continue
		} else {
			break
		}
	}
	return nil
}

func pingTCP(address string) error {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return errors.Wrap(err, "tcp ping failure")
	}
	defer conn.Close()
	return nil
}
