package service

import (
	"context"
	"fmt"
	grpcdialer "github.com/blademainer/commons/pkg/grpc"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	_ "github.com/pjoc-team/pay-gateway/pkg/config/file" // import config backend file
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	"github.com/pjoc-team/pay-gateway/pkg/util/network"
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
	_ "net/http/pprof" // import pprof
	"net/url"
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

const (
	// DefaultHTTPPort default http port of in process grpc gateway
	DefaultHTTPPort = 8080
	// DefaultHTTPGatewayPort use http gateway for grpc
	DefaultHTTPGatewayPort = 8088
	// DefaultInternalHTTPPort default internal http port
	DefaultInternalHTTPPort = 8081
	// DefaultGRPCPort default grpc port
	DefaultGRPCPort = 9090
	// DefaultPPROFPort default pprof port
	DefaultPPROFPort = 61616
)

// Server defined server
type Server struct {
	Ctx     context.Context
	FlagSet *pflag.FlagSet

	o                 *options
	g                 *errgroup.Group
	shutdownFunctions []ShutdownFunction
	services          *discovery.Services
	cancel            func()
}

type options struct {
	listen            int
	listenHTTP        int
	listenHTTPGateway int
	listenInternal    int
	listenPPROF       int
	network           string
	logLevel          string

	name              string
	infos             []*GrpcInfo
	shutdownFunctions []ShutdownFunction
	flagSet           []*pflag.FlagSet
	store             string
	enablePprof       bool
	// inProcessGateway  bool
}

func (o *options) apply(options ...Option) {
	for _, option := range options {
		option.apply(o)
	}
}

// Option service option
type Option interface {
	apply(opts *options)
}

// OptionFunc apply func
type OptionFunc func(*options)

func (o OptionFunc) apply(opts *options) {
	o(opts)
}

// ShutdownFunction shutdown func
type ShutdownFunction func(ctx context.Context)

// WithShutdown 增加关闭函数
func WithShutdown(function ShutdownFunction) Option {
	return OptionFunc(
		func(o *options) {
			o.shutdownFunctions = append(o.shutdownFunctions, function)
		},
	)
}

// WithGrpc 增加grpc服务
func WithGrpc(info *GrpcInfo) Option {
	return OptionFunc(
		func(o *options) {
			o.infos = append(o.infos, info)
		},
	)
}

// WithFlagSet add flagset
func WithFlagSet(flagSet *pflag.FlagSet) Option {
	return OptionFunc(
		func(o *options) {
			o.flagSet = append(o.flagSet, flagSet)
		},
	)
}

// func WithFlagSet(flagSet *pflag.FlagSet) Option {
//	return func(o *options) {
//		o.flagSet = append(o.flagSet, flagSet)
//	}
// }

// NewServer create server
func NewServer(name string, infos ...*GrpcInfo) (*Server, error) {
	log := logger.Log()
	err2 := logger.MinReportCallerLevel(logger.DebugLevel)
	if err2 != nil {
		log.Fatalf(err2.Error())
	}

	o := &options{
		name:  name,
		infos: infos,
	}
	s := &Server{
		o: o,
		services: &discovery.Services{},
	}
	fs := s.flags()
	s.FlagSet = fs
	rootContext, cancel := context.WithCancel(context.Background())
	s.Ctx = rootContext
	s.cancel = cancel
	return s, nil
}

func wordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	from := []string{"_", "."}
	to := "-"
	for _, sep := range from {
		name = strings.Replace(name, sep, to, -1)
	}
	return pflag.NormalizedName(name)
}

// GetDiscoveryServices get discovery services
func (s *Server) GetDiscoveryServices() *discovery.Services {
	return s.services
}

func (s *Server) initServices() (*discovery.Services, error) {
	store, err := discovery.NewFileStore(s.Ctx, s.o.store)
	if err != nil {
		logger.Log().Errorf(
			"failed to init file store of file: %v, error: %v", s.o.store, err.Error(),
		)
	}
	disc, err := discovery.NewDiscovery(store)
	if err != nil {
		logger.Log().Errorf(
			"failed to init file store of file: %v, error: %v", s.o.store, err.Error(),
		)
	}
	services := discovery.NewServices(disc)
	return services, nil
}

func (s *Server) flags() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("service", pflag.PanicOnError)
	flagSet.SetNormalizeFunc(wordSepNormalizeFunc)
	flagSet.IntVar(&s.o.listen, "listen", DefaultGRPCPort, "listen of the gRPC service")
	flagSet.IntVar(&s.o.listenHTTP, "listen-http", DefaultHTTPPort, "listen of the http service")
	flagSet.IntVar(
		&s.o.listenHTTPGateway, "listen-http-gateway", DefaultHTTPGatewayPort,
		"listen of the http grpc gateway",
	)
	flagSet.IntVar(
		&s.o.listenInternal, "listen-internal", DefaultInternalHTTPPort,
		"listen of the internal http service",
	)
	flagSet.IntVar(
		&s.o.listenPPROF, "listen-pprof", DefaultPPROFPort,
		"listen of the pprof http service",
	)
	flagSet.StringVar(&s.o.network, "network", "tcp", "network ")
	flagSet.StringVar(&s.o.logLevel, "log-level", "debug", "log level")
	flagSet.StringVar(&s.o.store, "store", "./conf/discovery.json", "file to store services")
	flagSet.BoolVar(
		&s.o.enablePprof, "enable-pprof", true,
		"turn on pprof debug tools",
	)
	for _, p := range s.o.flagSet {
		flagSet.AddFlagSet(p)
	}
	return flagSet
}

// Start start server
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

func (s Server) initDebug() {
	log := logger.Log()

	if s.o.enablePprof {
		// pprofFile := fmt.Sprintf("%s-cpu.prof", s.o.name)
		// f, err := os.Create(pprofFile)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// runtime.SetCPUProfileRate(10)
		// log.Infof("starting cpu profile to file: %v", pprofFile)
		// err = runtimepprof.StartCPUProfile(f)
		// if err != nil {
		// 	log.Errorf("failed to start cpu profile, error: %v", err.Error())
		// }
		// s.shutdownFunctions = append(
		// 	s.shutdownFunctions, func(ctx context.Context) {
		// 		runtimepprof.StopCPUProfile()
		// 		log.Warn("cpu profile is stopped")
		// 	},
		// )
		go func() {
			log.Warn("listening :61616 for pprof")
			err3 := http.ListenAndServe(fmt.Sprintf(":%d", s.o.listenPPROF), nil)
			if err3 != nil {
				log.Error("failed to listen pprof, error: %v", err3.Error())
			}
		}()
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
	defer s.cancel()
	g, ctx := errgroup.WithContext(s.Ctx)
	s.g = g

	err2 := s.initGrpc()
	if err2 != nil {
		log.Errorf("failed to init grpc, error: %v", err2.Error())
		log.Fatal(err2.Error())
	}

	s.initDebug()

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
		go shutdown(timeout)
	}

	s.cancel()
	err := g.Wait()
	if err != nil {
		log.Errorf("server returning an error: %v", err)
	}
}

// InitLoggerAndTracing init logger and tracing
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

func (s *Server) initGrpc() error {
	g := s.g
	ctx := s.Ctx
	log := logger.ContextLog(ctx)

	services, err := s.initServices()
	if err != nil {
		log.Errorf("failed to init services, error: %v", err.Error())
		return err
	}

	*s.services = *services

	ip, err := network.GetHostIP()
	if err != nil {
		log.Errorf("failed to get host ip, error: %v", err.Error())
		return err
	}
	for _, info := range s.o.infos {
		// 注册所有grpc
		RegisterGrpc(
			info.Name, info.RegisterGrpcFunc, info.RegisterGatewayFunc, info.RegisterStreamFunc,
		)
		err := s.services.Discovery.RegisterService(
			info.Name, &discovery.Service{
				ServiceName: info.Name,
				Protocol:    discovery.GRPC,
				IP:          ip,
				Port:        s.o.listen,
			},
		)

		if err != nil {
			log.Errorf("failed to register grpc: %v, eror: %v", info.Name, err.Error())
			return err
		}
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
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_prometheus.UnaryServerInterceptor,
				tracinggrpc.TracingServerInterceptor(), // tracing
				grpc_recovery.UnaryServerInterceptor(opts...),
			),
		),
	)

	// init grpc
	g.Go(
		func() error {
			log.Infof("grpc listen %v", s.o.listen)
			l, err := net.Listen(s.o.network, fmt.Sprintf(":%d", s.o.listen))
			if err != nil {
				return err
			}

			// register health server
			healthpb.RegisterHealthServer(gs, healthServer)

			// register services
			for k, registerGrpc := range GrpcServices {
				log.Infof("initializing grpc: %v", k)
				err := registerGrpc.RegisterGrpcFunc(s.Ctx, gs)
				if err != nil {
					log.Fatalf("failed to register grpc: %v error: %v", k, err.Error())
				} else {
					log.Infof("succeed register grpc: %v", k)
				}
			}

			for k, serviceInfo := range gs.GetServiceInfo() {
				// logger.Infof("services name: %v info: %v", k, serviceInfo.Metadata, serviceInfo.Methods)
				for _, method := range serviceInfo.Methods {
					log.Infof(
						"services name: %v info: %v method: %v", k, serviceInfo.Metadata,
						method.Name,
					)
				}
			}

			grpc_prometheus.Register(gs)

			s.shutdownFunctions = append(
				s.shutdownFunctions, func(ctx context.Context) {
					gs.GracefulStop()

					// if err := l.Close(); err != nil {
					// 	log.Errorf("failed to close %v %v, err: %v", s.o.network, s.o.listen, err)
					// }
				},
			)

			return gs.Serve(l)
		},
	)

	// http admin
	g.Go(
		func() error {
			log.Infof("admin listen %v", s.o.listenInternal)
			listen, err := net.Listen("tcp", fmt.Sprintf(":%d", s.o.listenInternal))
			if err != nil {
				log.Errorf("failed to listen: %v error: %v", s.o.listenInternal, err.Error())
				return err
			}

			internalHTTPMux.Handle("/metrics", promhttp.Handler())
			h := healthInterceptor(healthServer)
			internalHTTPMux.Handle("/health", h)
			// pprof
			// if log.IsDebugEnabled() {
			// 	internalHTTPMux.HandleFunc("/debug/pprof/", pprof.Index)
			// }

			httpServer := &http.Server{
				Addr:         fmt.Sprintf(":%d", s.o.listenInternal),
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
				Handler:      tracingServerInterceptor(internalHTTPMux),
			}
			s.shutdownFunctions = append(
				s.shutdownFunctions, func(ctx context.Context) {
					err2 := httpServer.Shutdown(ctx)
					if err2 != nil && err2 != http.ErrServerClosed {
						log.Errorf(
							"failed to close: %v error: %v", s.o.listenInternal, err2.Error(),
						)
						return
					}
					log.Infof("http admin closed")
				},
			)

			err = httpServer.Serve(listen)
			if err != http.ErrServerClosed {
				return err
			}
			log.Infof("http stopped")
			return nil
		},
	)

	// in processor grpc gateway
	g.Go(
		func() error {
			mux := newGrpcMux()

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

			h := intercept(mux)

			hs := &http.Server{
				Addr:    fmt.Sprintf(":%d", s.o.listenHTTP),
				Handler: h,
			}

			s.shutdownFunctions = append(
				s.shutdownFunctions, func(ctx context.Context) {
					if err := hs.Shutdown(context.Background()); err != nil {
						log.Errorf("Failed to shutdown http gateway server: %v", err)
					}
				},
			)

			log.Infof("grpc gateway listen %v", s.o.listenHTTP)
			if err := hs.ListenAndServe(); err != http.ErrServerClosed {
				log.Errorf("Failed to listen and serve: %v", err)
				return err
			}

			return nil
		},
	)

	// http grpc gateway
	g.Go(
		func() error {
			log.Infof("grpc http gateway listen %v", s.o.listenHTTPGateway)
			grpcURL, err := url.Parse(fmt.Sprintf("http://localhost:%d", s.o.listen))
			if err != nil {
				log.Fatal(err.Error())
			}
			conn, err := grpcdialer.DialUrl(ctx, *grpcURL)
			if err != nil {
				log.Fatal(err.Error())
			}
			s.shutdownFunctions = append(
				s.shutdownFunctions, func(ctx context.Context) {
					log.Infof("closing client connection")
					err2 := conn.Close()
					if err2 != nil {
						log.Errorf(
							"Failed to close a client connection to the gRPC server: %v",
							err,
						)
					}
					log.Infof("grpc client closed")
				},
			)

			grpcGatewayMux := newGrpcMux()

			h := intercept(grpcGatewayMux)

			for k, registerGrpc := range GrpcServices {
				if registerGrpc.RegisterStreamFunc == nil {
					log.Warnf("registerGrpc: %v's streaming is nil", registerGrpc)
					continue
				}
				log.Infof("initializing grpc streaming: %v", k)
				err := registerGrpc.RegisterStreamFunc(ctx, grpcGatewayMux, conn)
				if err != nil {
					log.Fatalf("failed to register grpc: %v error: %v", k, err.Error())
				} else {
					log.Infof("succeed register grpc gateway: %v", k)
				}
			}
			httpMux := http.NewServeMux()
			httpMux.Handle("/", h)
			hs := &http.Server{
				Addr:    fmt.Sprintf(":%d", s.o.listenHTTPGateway),
				Handler: httpMux,
			}

			s.shutdownFunctions = append(
				s.shutdownFunctions, func(ctx context.Context) {
					err2 := hs.Shutdown(context.Background())
					if err2 != nil {
						log.Error(err2.Error())
					}
				},
			)
			err2 := hs.ListenAndServe()
			if err2 != http.ErrServerClosed {
				return err2
			}
			log.Infof("grpc http gateway closed")
			return nil
		},
	)
	return nil
}
