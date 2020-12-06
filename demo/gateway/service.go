package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/pjoc-team/pay-gateway/demo/gateway/proto"
	"github.com/pjoc-team/pay-gateway/demo/gateway/stream"
	"google.golang.org/grpc"
	"net"
	"net/http"
)

var (
	endpoint = flag.String("endpoint", "localhost:9090", "endpoint of the gRPC service")
	network  = flag.String(
		"network", "tcp", `one of "tcp" or "unix". Must be consistent to -endpoint`,
	)
	openAPIDir = flag.String(
		"openapi_dir", "examples/internal/proto/examplepb",
		"path to the directory which contains OpenAPI definitions",
	)
)

func runGateway(ctx context.Context, addr string, opts ...runtime.ServeMuxOption) error {
	return Run(
		ctx, Options{
			Addr: addr,
			GRPCServer: Endpoint{
				Network: *network,
				Addr:    *endpoint,
			},
			OpenAPIDir: *openAPIDir,
			Mux:        opts,
		},
	)
}

func main() {
	flag.Parse()
	ctx, cancelFunc := context.WithCancel(context.TODO())
	defer cancelFunc()
	go func() {
		err := RunGrpc(ctx, "tcp", ":9090")
		if err != nil {
			panic(err.Error())
		}
	}()
	if err := runGateway(ctx, ":8089"); err != nil {
		panic(err.Error())
	}
}

// Endpoint describes a gRPC endpoint
type Endpoint struct {
	Network, Addr string
}

// Options is a set of options to be passed to Run
type Options struct {
	// Addr is the address to listen
	Addr string

	// GRPCServer defines an endpoint of a gRPC service
	GRPCServer Endpoint

	// OpenAPIDir is a path to a directory from which the server
	// serves OpenAPI specs.
	OpenAPIDir string

	// Mux is a list of options to be passed to the grpc-gateway multiplexer
	Mux []runtime.ServeMuxOption
}

// Run starts a HTTP server and blocks while running if successful.
// The server will be shutdown when "ctx" is canceled.
func Run(ctx context.Context, opts Options) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := dial(ctx, opts.GRPCServer.Network, opts.GRPCServer.Addr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			glog.Errorf("Failed to close a client connection to the gRPC server: %v", err)
		}
	}()

	mux := http.NewServeMux()


	gw, err := newGateway(ctx, conn, opts.Mux)
	if err != nil {
		return err
	}
	mux.Handle("/", gw)

	s := &http.Server{
		Addr:    opts.Addr,
		Handler: mux,
	}
	go func() {
		<-ctx.Done()
		glog.Infof("Shutting down the http server")
		if err := s.Shutdown(context.Background()); err != nil {
			glog.Errorf("Failed to shutdown http server: %v", err)
		}
	}()

	glog.Infof("Starting listening at %s", opts.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		glog.Errorf("Failed to listen and serve: %v", err)
		return err
	}
	return nil
}

// Run starts the example gRPC service.
// "network" and "address" are passed to net.Listen.
func RunGrpc(ctx context.Context, network, address string) error {
	l, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	defer func() {
		if err := l.Close(); err != nil {
			glog.Errorf("Failed to close %s %s: %v", network, address, err)
		}
	}()

	s := grpc.NewServer()

	ss := stream.NewStreamServer()
	pb.RegisterStreamServiceServer(s, ss)

	go func() {
		defer s.GracefulStop()
		<-ctx.Done()
	}()
	return s.Serve(l)
}

// newGateway returns a new gateway server which translates HTTP into gRPC.
func newGateway(
	ctx context.Context, conn *grpc.ClientConn, opts []runtime.ServeMuxOption,
) (http.Handler, error) {

	mux := runtime.NewServeMux(opts...)
	// runtime.SetHTTPBodyMarshaler(mux)

	for _, f := range []func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error{
		pb.RegisterStreamServiceHandler,
	} {
		if err := f(ctx, mux, conn); err != nil {
			return nil, err
		}
	}
	return mux, nil
}

func dial(ctx context.Context, network, addr string) (*grpc.ClientConn, error) {
	switch network {
	case "tcp":
		return dialTCP(ctx, addr)
	case "unix":
		return dialUnix(ctx, addr)
	default:
		return nil, fmt.Errorf("unsupported network type %q", network)
	}
}

// dialTCP creates a client connection via TCP.
// "addr" must be a valid TCP address with a port number.
func dialTCP(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, addr, grpc.WithInsecure())
}

// dialUnix creates a client connection via a unix domain socket.
// "addr" must be a valid path to the socket.
func dialUnix(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	d := func(ctx context.Context, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", addr)
	}
	return grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithContextDialer(d))
}
