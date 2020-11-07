package service

import (
	"encoding/json"
	"github.com/pjoc-team/tracing/logger"
)

// Discovery discovery server
type Discovery struct {
}

// Service define service entity
type Service struct {
	ServiceName string `json:"service_name" yaml:"serviceName"`
	Protocol    string `json:"protocol" yaml:"protocol"`
	IP          string `json:"ip" yaml:"ip"`
	Port        int    `json:"port" yaml:"port"`
}

// GetService discovery service
func (d *Discovery) GetService(serviceName string) (string, error) {
	return "", nil
}

// GetService discovery service
func (d *Discovery) RegisterService(serviceName string) (string, error) {
	return "", nil
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
