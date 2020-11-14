package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pjoc-team/tracing/logger"
	"strings"
)

// Protocol service protocol
type Protocol string

const (
	GRPC Protocol = "grpc"
	Http Protocol = "http"
)

// StoreIsNil init failed
var StoreIsNil = errors.New("store is nil")

// Discovery discovery server
type Discovery struct {
	store Store
}

// NewDiscovery new discovery server
func NewDiscovery(store Store) (*Discovery, error) {
	if store == nil {
		return nil, StoreIsNil
	}
	d := &Discovery{
		store: store,
	}
	return d, nil
}

// Service define service entity
type Service struct {
	ServiceName string   `json:"service_name" yaml:"serviceName"`
	Protocol    Protocol `json:"protocol" yaml:"protocol"`
	IP          string   `json:"ip" yaml:"ip"`
	Port        int      `json:"port" yaml:"port"`
}

// String stringer
func (s *Service) String() string {
	builder := strings.Builder{}

	builder.WriteString("ServiceName=")
	builder.WriteString(s.ServiceName)
	builder.WriteString(",Protocol=")
	builder.WriteString(string(s.Protocol))
	builder.WriteString(",IP=")
	builder.WriteString(s.IP)
	builder.WriteString(",Port=")
	builder.WriteString(string(s.Port))

	return builder.String()
}

// BuildTarget build rpc connection target
func (s *Service) BuildTarget(ctx context.Context) (string, error) {
	log := logger.ContextLog(ctx)
	target := fmt.Sprintf("%s:%d", s.ServiceName, s.Port)
	if log.IsDebugEnabled() {
		log.Debugf("build service: %#v to target: %v", s, target)
	}
	return target, nil
}

// GetService discovery service
func (d *Discovery) GetService(ctx context.Context, serviceName string) (*Service, error) {
	log := logger.ContextLog(ctx)
	service, err := d.store.Get(serviceName)
	if err != nil { // not registered
		log.Warnf("not found service: %v, so use serviceName and default port", serviceName)
		s := &Service{
			ServiceName: serviceName,
			Protocol:    GRPC,
			IP:          "",
			Port:        9090,
		}
		return s, nil
	}
	return service, nil
}

// GetService discovery service
func (d *Discovery) RegisterService(serviceName string, service *Service) error {
	err := d.store.Put(serviceName, service)
	return err
}

// Marshal marshal to raw
func (s Service) Marshal() (string, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		logger.Log().Errorf("failed to marshal service: %#v error: %v", s, err.Error())
		return "", err
	}
	return string(raw), nil
}

// Unmarshal unmarshal string to service
func Unmarshal(raw string) (*Service, error) {
	s := &Service{}
	err := json.Unmarshal([]byte(raw), s)
	if err != nil {
		logger.Log().Errorf("failed to unmarshal service: %#v error: %v", raw, err.Error())
		return nil, err
	}
	return s, nil
}
