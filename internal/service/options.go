package service

import (
	"context"
	"github.com/spf13/pflag"
)

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
