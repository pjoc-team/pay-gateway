package config

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/pjoc-team/pay-gateway/pkg/config/types"
	tracinglogger "github.com/pjoc-team/tracing/logger"
)

//const (
//	urlPatternStr = "\\w+://(\\w+:\\w+@)*(\\w+.)*\\w+(\\?(\\w+=\\w+&)(\\w+=\\w+))*"
//)

// Server 配置服务
type Server interface {
	// GetConfig 获取配置并设置到ptr，keys是树形结构
	UnmarshalGetConfig(ctx context.Context, ptr interface{}, keys ...string) error
}

type defaultServer struct {
	config  *types.Config
	backend types.Backend
}

func InitConfigServer(urlStr string) (Server, error) {
	s := &defaultServer{}

	config, err := types.ParseConfig(urlStr)
	if err != nil {
		return nil, err
	}
	s.config = config

	bf, err := types.GetBackend(types.Provider(config.BackendType))
	if err != nil {
		return nil, err
	}
	backend, err := bf.InitFunc(config)
	if err != nil {
		err = fmt.Errorf("failed to init backend, with error: %v a demo config url: %v", err.Error(), bf.Options.DemoURL)
		return nil, err
	}
	v := validator.New()
	err = v.Struct(backend)
	if err != nil {
		panic(fmt.Sprintf("validate backend config error: %v a correct url is: '%v'", err.Error(), bf.Options.DemoURL))
	}

	s.backend = backend

	return s, nil
}

// GetConfig 获取配置，将会把配置放置到ptr内，其中keys是主键（可以是多个）
func (s *defaultServer) UnmarshalGetConfig(ctx context.Context, ptr interface{}, keys ...string) error {
	err := s.backend.UnmarshalGetConfig(ctx, ptr, keys...)
	log := tracinglogger.ContextLog(ctx)
	if err != nil {
		log.Errorf("failed to get config: %v error: %v", keys, err.Error())
	} else {
		log.Debugf("get config: %#v by keys: %v", ptr, keys)
	}
	return err
}
