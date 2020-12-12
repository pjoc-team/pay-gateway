package discovery

import (
	"context"
	"errors"
	grpcdialer "github.com/blademainer/commons/pkg/grpc"
	"github.com/pjoc-team/tracing/logger"
	"github.com/pjoc-team/tracing/tracinggrpc"
	"google.golang.org/grpc"
	"net/url"
	"sync"
	"time"
)

// ErrGetConnectionFailed get connection error
var ErrGetConnectionFailed = errors.New("failded to get connection")

// ServiceName service name
type ServiceName string

func (s ServiceName) String() string {
	return string(s)
}

// Services defined services
type Services struct {
	Discovery   *Discovery
	rwLocker    sync.RWMutex
	servicePool map[string]*sync.Pool
}

// NewServices create service discovery
func NewServices(discovery *Discovery) *Services {
	s := &Services{
		Discovery:   discovery,
		rwLocker:    sync.RWMutex{},
		servicePool: make(map[string]*sync.Pool),
	}
	return s
}

// PutBackClientFunc put grpc connection back
type PutBackClientFunc func()

// PutClient put client return to the pool
func (s *Services) PutClient(serviceName string, conn *grpc.ClientConn) {
	pool := s.getPool(serviceName)
	pool.Put(conn)
}

func (s *Services) getPool(serviceName string) *sync.Pool {
	s.rwLocker.RLock()
	defer s.rwLocker.RUnlock()
	pool := s.servicePool[serviceName]
	if pool != nil {
		return pool
	}
	// try to get lock
	s.rwLocker.RUnlock()
	defer s.rwLocker.RLock()
	s.rwLocker.Lock()
	defer s.rwLocker.Unlock()
	pool = s.servicePool[serviceName]
	if pool != nil {
		return pool
	}
	pool = &sync.Pool{
		New: func() interface{} {
			return s.dialService(serviceName)
		},
	}
	s.servicePool[serviceName] = pool
	return pool
}

func (s *Services) initGrpc(
	ctx context.Context, serviceName string,
	grpcFunc func(conn *grpc.ClientConn) interface{},
) (interface{}, PutBackClientFunc, error) {
	log := logger.ContextLog(ctx)
	pool := s.getPool(serviceName)
	d := pool.Get()
	if d == nil {
		log.Error(ErrGetConnectionFailed.Error())
		return nil, nil, ErrGetConnectionFailed
	}
	conn := d.(*grpc.ClientConn)
	client := grpcFunc(conn)
	return client, func() { pool.Put(d) }, nil
}

// DialTarget dial grpc target
func DialTarget(ctx context.Context, target string) (*grpc.ClientConn, error) {
	log := logger.ContextLog(ctx)
	u, err := url.Parse(target)
	if err != nil {
		log.Errorf(
			"failed to build target: %v, error: %v", target, err.Error(),
		)
		return nil, err
	}
	d, err := grpcdialer.DialUrl(
		ctx, *u, grpc.WithChainUnaryInterceptor(
			tracinggrpc.TracingClientInterceptor(),
		),
	)
	return d, err
}

func (s *Services) dialService(serviceName string) *grpc.ClientConn {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log := logger.Log()
	svc, err := s.Discovery.GetService(ctx, serviceName)
	if err != nil {
		log.Errorf("failed to get service: %v error: %v", serviceName, err.Error())
		return nil
	}
	if svc.Protocol != "" && svc.Protocol != GRPC {
		log.Errorf(
			"service: %v's protocol is not grpc, "+
				"actual is: %v and continue try to connect by grpc protocol",
			serviceName,
			svc.Protocol,
		)
	}

	target, err := svc.BuildTarget(ctx)
	if err != nil {
		log.Errorf("failed to build target of service: %v error: %v", serviceName, err.Error())
		return nil
	}

	d, err := DialTarget(ctx, target)

	if err != nil {
		return nil
	}
	return d
}
