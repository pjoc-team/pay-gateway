package types

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/pjoc-team/tracing/logger"
	"net/url"
	"reflect"
)

// Options 设置
type Options struct {
	TestDemoURL bool
	DemoURL     string
}

func (o *Options) apply(options ...Option) {
	for _, option := range options {
		option(o)
	}
}

type Option func(*Options)

// TestDemoURL 测试url
func TestDemoURL() Option {
	return func(options *Options) {
		options.TestDemoURL = true
	}
}

// WithDemoURL 传入注册url
func WithDemoURL(demoURL string) Option {
	return func(options *Options) {
		options.DemoURL = demoURL
	}
}

// Config 配置，由URL反解析而来
type Config struct {
	BackendUrl  *url.URL
	BackendType string
	Path        string
	Args        map[string][]string
	User        *url.Userinfo
	Host        string
}

// InitFunc 初始化函数，管理器会在初始化的时候使用validator来校验backend是否合法，配置方法参考：https://godoc.org/gopkg.in/go-playground/validator.v10
type InitFunc func(config *Config) (Backend, error)

//go:generate mockgen -source ./backend.go -package mock -destination ./mock/mock.go

// Backend 后端实现
type Backend interface {
	// Start 开始监听配置
	Start() error

	// GetConfig 获取配置并设置到ptr，keys是树形结构
	UnmarshalGetConfig(ctx context.Context, ptr interface{}, keys ...string) error
}

// RegisterBackendOrDie 注册后端实现
func RegisterBackendOrDie(provider Provider, initFunc InitFunc, options ...Option) error {
	log := logger.Log()

	_, ok := backends[provider]
	if ok {
		panic(fmt.Sprintf("provider: %v is already registered, func: %v", provider, reflect.ValueOf(initFunc)))
	}
	opts := &Options{}
	opts.apply(options...)

	if opts.DemoURL != "" {
		config, err := ParseConfig(opts.DemoURL)
		if err != nil {
			panic(fmt.Sprintf("url demo is illegal, url:%v error: %v", opts.DemoURL, err.Error()))
		}
		if opts.TestDemoURL {
			backend, err := initFunc(config)
			if err != nil {
				panic(fmt.Sprintf("failed to register backend, url:%v error: %v", opts.DemoURL, err.Error()))
			}
			vld := validator.New()
			err = vld.Struct(backend)
			if err != nil {
				panic(fmt.Sprintf("validate backend error: %v", err.Error()))
			}
		}
	} else if opts.TestDemoURL {
		log.Warn("TestDemoURL is on but no DemoURL is present")
	}

	backends[provider] = &InitOpts{
		InitFunc: initFunc,
		Options:  opts,
	}
	return nil
}

// ParseConfig 解析url，例如file://foo/bar/config.yaml
func ParseConfig(urlStr string) (*Config, error) {
	c := &Config{}
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	c.BackendUrl = u
	c.BackendType = u.Scheme
	c.Host = u.Host
	c.Path = u.Path
	c.Args, err = url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}
	c.User = u.User

	return c, nil
}
