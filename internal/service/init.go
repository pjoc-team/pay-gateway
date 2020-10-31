package service

import (
	"context"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	_ "github.com/pjoc-team/pay-gateway/pkg/config/file"
	"github.com/pjoc-team/pay-gateway/pkg/metadata"
	"github.com/pjoc-team/tracing/logger"
	"github.com/pjoc-team/tracing/tracing"
	"github.com/pjoc-team/tracing/tracinggrpc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"math/rand"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// var (
//	listen         = flag.String("listen", ":9090", "listen of the gRPC service")
//	listenHTTP     = flag.String("listenHTTP", ":8080", "listen of the http service")
//	listenInternal = flag.String("listenInternal", ":8081", "listen of the internal http service")
//	network        = flag.String("network", "tcp", "network ")
//	logLevel       = flag.String("log-level", "debug", "log level")
// )

type Server struct {
	o                 *options
	ctx               context.Context
	g                 *errgroup.Group
	shutdownFunctions []ShutdownFunction
}

type options struct {
	listen         string
	listenHTTP     string
	listenInternal string
	network        string
	logLevel       string

	name              string
	infos             []*GrpcInfo
	shutdownFunctions []ShutdownFunction
	flagSet           []*pflag.FlagSet
}

func (o *options) apply(options ...Option) {
	for _, option := range options {
		option(o)
	}
}

type Option func(*options)

type ShutdownFunction func(ctx context.Context)

// WithShutdown 增加关闭函数
func WithShutdown(function ShutdownFunction) Option {
	return func(o *options) {
		o.shutdownFunctions = append(o.shutdownFunctions, function)
	}
}

// WithGrpc 增加grpc服务
func WithGrpc(info *GrpcInfo) Option {
	return func(o *options) {
		o.infos = append(o.infos, info)
	}
}

// func WithFlagSet(flagSet *pflag.FlagSet) Option {
//	return func(o *options) {
//		o.flagSet = append(o.flagSet, flagSet)
//	}
// }

func NewServer(name string, infos ...*GrpcInfo) (*Server, *pflag.FlagSet, error) {
	o := &options{
		name:  name,
		infos: infos,
	}
	s := &Server{
		o: o,
	}
	fs := s.flags()
	return s, fs, nil
}

func wordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	from := []string{"_", "."}
	to := "-"
	for _, sep := range from {
		name = strings.Replace(name, sep, to, -1)
	}
	return pflag.NormalizedName(name)
}

func (s *Server) flags() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("service", pflag.PanicOnError)
	flagSet.SetNormalizeFunc(wordSepNormalizeFunc)
	flagSet.StringVar(&s.o.listen, "listen", ":9090", "listen of the gRPC service")
	flagSet.StringVar(&s.o.listenHTTP, "listen-http", ":8080", "listen of the http service")
	flagSet.StringVar(&s.o.listenInternal, "listen-internal", ":8081", "listen of the internal http service")
	flagSet.StringVar(&s.o.network, "network", "tcp", "network ")
	flagSet.StringVar(&s.o.logLevel, "log-level", "debug", "log level")
	for _, p := range s.o.flagSet {
		flagSet.AddFlagSet(p)
	}
	return flagSet
}

func (s *Server) Start(opts ...Option) {
	rand.Seed(int64(time.Now().Nanosecond()))
	s.o.apply(opts...)
	flags := s.flags()
	cmd := cobra.Command{
		Use:                "",
		Short:              "",
		Long:               "",
		DisableFlagParsing: true,
		Run:                s.runFunc(flags),
	}
	if err := cmd.Execute(); err != nil {
		logger.Log().Fatal(err.Error())
	}
}

func (s *Server) runFunc(flagSet *pflag.FlagSet) func(cmd *cobra.Command, args []string) {
	log := logger.Log()
	return func(cmd *cobra.Command, args []string) {
		if err := flagSet.Parse(args); err != nil {
			err2 := cmd.Usage()
			if err2 != nil {
				log.Error(err2.Error())
			}

			log.Fatal(err.Error())
		}

		s.InitLoggerAndTracing(s.o.name)

		// check if there are non-flag arguments in the command line
		cmds := flagSet.Args()

		if len(cmds) > 0 {
			err2 := cmd.Usage()
			if err2 != nil {
				log.Error(err2.Error())
			}
			log.Fatalf("unknown command: %s", cmds[0])
		}
		log.Infof("flags: %v", flagSet)

		// short-circuit on help
		help, _ := flagSet.GetBool("help")
		// if err != nil {
		//	logger.Fatalf(`"help" flag is non-bool, programmer error, please correct. error: %v`, err.Error())
		// }
		if help {
			err2 := cmd.Help()
			if err2 != nil {
				log.Error(err2.Error())
			}
			return
		}
		s.run()
	}
}

func (s *Server) run() {
	log := logger.Log()
	if len(s.o.infos) == 0 {
		log.Fatal("no grpc infos")
	}

	// shutdown functions
	s.shutdownFunctions = make([]ShutdownFunction, 0)
	s.shutdownFunctions = append(s.shutdownFunctions, s.o.shutdownFunctions...)

	// init
	rootContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(rootContext)
	s.ctx = ctx
	s.g = g

	s.initGrpc()

	// signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	log.Infof("waiting for interrupt...")
	select {
	case <-ctx.Done():
		log.Warnf("timeout...")
	case <-interrupt:
		log.Warnf("interrupt received!!!")
	}

	// 创建一个新的Context，等待各个服务释放资源
	timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	for _, shutdown := range s.shutdownFunctions {
		shutdown(timeout)
	}

	cancel()
	err := g.Wait()
	if err != nil {
		log.Errorf("server returning an error: %v", err)
	}
}

func (s *Server) InitLoggerAndTracing(serviceName string) {
	// setting logger
	log := logger.Log()
	level, err2 := logger.ParseLevel(s.o.logLevel)
	if err2 != nil {
		log.Fatalf("failed to setting level: %v", err2.Error())
	}
	log.Infof("log level: %v", s.o.logLevel)
	err2 = logger.SetLevel(level)
	if err2 != nil {
		log.Fatalf("failed to setting level: %v", err2.Error())
	}
	// reset after settings log level
	log = logger.Log()
	// setting logger to grpc
	grpclog.SetLoggerV2(log)
	err2 = os.Setenv("JAEGER_SERVICE_NAME", serviceName)
	if err2 != nil {
		log.Fatalf("failed to init tracing, error: %v", err2.Error())
	}
	err2 = tracing.InitFromEnv()
	if err2 != nil {
		log.Fatalf("failed to init tracing, error: %v", err2.Error())
	}
	// err2 = tracing.InitOnlyTracingLog("drive")
	// if err2 != nil {
	//	logger.Fatalf(err2.Error())
	// }
	// 需要打印调用来源的日志级别
	err2 = logger.MinReportCallerLevel(level)
	if err2 != nil {
		log.Fatalf(err2.Error())
	}
}

func (s *Server) initGrpc() {
	g := s.g
	ctx := s.ctx
	log := logger.Log()
	for _, info := range s.o.infos {
		// 注册所有grpc
		RegisterGrpc(info.Name, info.RegisterGrpcFunc, info.RegisterGatewayFunc)
	}

	// grpc server
	// health service
	healthServer := health.NewServer()
	internalHTTPMux := http.NewServeMux() // 内部服务的http

	// 自定义错误处理
	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandlerContext(customRecoverFunc),
	}

	// init grpc server
	gs := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			tracinggrpc.TracingServerInterceptor(), // tracing
			grpc_recovery.UnaryServerInterceptor(opts...),
		)),
	)

	// init grpc gateway
	marshaler := &runtime.JSONPb{
		EnumsAsInts:  false, // 枚举类使用string返回
		OrigName:     true,  // 使用json tag里面的字段
		EmitDefaults: true,  // json返回零值
	}
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaler),
		runtime.WithMetadata(metadata.ParseHeaderAndQueryToMD),
		runtime.WithProtoErrorHandler(protoErrorHandler),
	)

	// init grpc
	g.Go(func() error {
		log.Infof("grpc listen %s", s.o.listen)
		l, err := net.Listen(s.o.network, s.o.listen)
		if err != nil {
			return err
		}

		// register health server
		healthpb.RegisterHealthServer(gs, healthServer)

		// register services
		for k, registerGrpc := range GrpcServices {
			log.Infof("initializing grpc: %v", k)
			err := registerGrpc.RegisterGrpcFunc(s.ctx, gs)
			if err != nil {
				log.Fatalf("failed to register grpc: %v error: %v", k, err.Error())
			} else {
				log.Infof("succeed register grpc: %v", k)
			}
		}

		for k, serviceInfo := range gs.GetServiceInfo() {
			// logger.Infof("services name: %v info: %v", k, serviceInfo.Metadata, serviceInfo.Methods)
			for _, method := range serviceInfo.Methods {
				log.Infof("services name: %v info: %v method: %v", k, serviceInfo.Metadata, method.Name)
			}
		}

		grpc_prometheus.Register(gs)

		s.shutdownFunctions = append(s.shutdownFunctions, func(ctx context.Context) {
			gs.GracefulStop()

			if err := l.Close(); err != nil {
				log.Errorf("Failed to close %s %s, err: %v", s.o.network, s.o.listen, err)
			}
		})

		return gs.Serve(l)
	})

	// http admin
	g.Go(func() error {
		log.Infof("admin listen %s", s.o.listenInternal)
		listen, err := net.Listen("tcp", s.o.listenInternal)
		if err != nil {
			log.Errorf("failed to listen: %v error: %v", s.o.listenInternal, err.Error())
			return err
		}
		s.shutdownFunctions = append(s.shutdownFunctions, func(ctx context.Context) {
			err2 := listen.Close()
			if err2 != nil {
				log.Errorf("failed to close: %v error: %v", s.o.listenInternal, err2.Error())
			} else {
				log.Infof("http admin closed")
			}
		})

		internalHTTPMux.Handle("/metrics", promhttp.Handler())
		h := healthInterceptor(healthServer)
		internalHTTPMux.Handle("/health", h)
		// pprof
		if log.IsDebugEnabled() {
			internalHTTPMux.HandleFunc("/debug/pprof/", pprof.Index)
		}

		httpServer := &http.Server{
			Addr:         s.o.listenInternal,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      tracingServerInterceptor(internalHTTPMux),
		}
		err = httpServer.Serve(listen)
		log.Infof("http stopped")
		return err
	})

	// grpc gateway
	g.Go(func() error {

		// register grpc gateway services
		for k, registerGrpc := range GrpcServices {
			log.Infof("initializing grpc: %v", k)
			err := registerGrpc.RegisterGatewayFunc(ctx, mux)
			if err != nil {
				log.Fatalf("failed to register grpc: %v error: %v", k, err.Error())
			} else {
				log.Infof("succeed register grpc gateway: %v", k)
			}
		}

		// 注入tracing信息
		var h http.Handler = mux
		for _, interceptor := range httpInterceptors {
			h = interceptor(h)
		}

		hs := &http.Server{
			Addr:    s.o.listenHTTP,
			Handler: h,
		}

		s.shutdownFunctions = append(s.shutdownFunctions, func(ctx context.Context) {
			if err := hs.Shutdown(context.Background()); err != nil {
				log.Errorf("Failed to shutdown http gateway server: %v", err)
			}
		})

		log.Infof("grpc gateway listen %s", s.o.listenHTTP)
		if err := hs.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Failed to listen and serve: %v", err)
			return err
		}

		return nil
	})

}
