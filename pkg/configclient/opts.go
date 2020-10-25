package configclient

import "github.com/spf13/pflag"

// options 配置选项
type options struct {
	ps                            *pflag.FlagSet
	PayConfigServerURL            PayConfigServerURL
	NoticeConfigServerURL         NoticeConfigServerURL
	ServiceConfigServerURL        ServiceConfigServerURL
	ChannelServiceConfigServerURL ChannelServiceConfigServerURL
	MerchantConfigServerURL       MerchantConfigServerURL
	PersonalMerchantServerURL     PersonalMerchantServerURL
	AppIdChannelConfigServerURL   AppIdChannelConfigServerURL
}

func newOpts() (*options, error) {
	set := pflag.NewFlagSet("config-url", pflag.PanicOnError)

	o := &options{
		ps: set,
	}
	return o, nil
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// Option 选项
type Option func(o *options)

// WithPayConfigServer 设置PayConfigServer
func WithPayConfigServer(required bool) Option {
	return func(o *options) {
		if !required {
			return
		}
		u := o.PayConfigServerURL
		s := string(o.PayConfigServerURL)
		o.ps.StringVar(&s, u.Flag(), "file://./conf/pay-config.yaml", "config uri. see: config.Server")
	}
}

// WithNoticeConfigServer 设置NoticeConfigServer
func WithNoticeConfigServer(required bool) Option {
	return func(o *options) {
		if !required {
			return
		}
		u := o.NoticeConfigServerURL
		s := string(o.NoticeConfigServerURL)
		o.ps.StringVar(&s, u.Flag(), "file://./conf/notice-config.yaml", "config uri. see: config.Server")
	}
}

// WithServiceConfigServer 设置ServiceConfigServer
func WithServiceConfigServer(required bool) Option {
	return func(o *options) {
		if !required {
			return
		}
		u := o.ServiceConfigServerURL
		s := string(o.ServiceConfigServerURL)
		o.ps.StringVar(&s, u.Flag(), "file://./conf/service-config.yaml", "config uri. see: config.Server")
	}
}

// WithChannelServiceConfigServer 设置ChannelServiceConfigServer
func WithChannelServiceConfigServer(required bool) Option {
	return func(o *options) {
		if !required {
			return
		}
		u := o.ChannelServiceConfigServerURL
		s := string(o.ChannelServiceConfigServerURL)
		o.ps.StringVar(&s, u.Flag(), "file://./conf/channel-service-config.yaml", "config uri. see: config.Server")
	}
}

// WithMerchantConfigServer 设置MerchantConfigServer
func WithMerchantConfigServer(required bool) Option {
	return func(o *options) {
		if !required {
			return
		}
		u := o.MerchantConfigServerURL
		s := string(o.MerchantConfigServerURL)
		o.ps.StringVar(&s, u.Flag(), "file://./conf/merchant-config.yaml", "config uri. see: config.Server")
	}
}

// WithPersonalMerchantServer 设置PersonalMerchantServer
func WithPersonalMerchantServer(required bool) Option {
	return func(o *options) {
		if !required {
			return
		}
		u := o.PersonalMerchantServerURL
		s := string(o.PersonalMerchantServerURL)
		o.ps.StringVar(&s, u.Flag(), "file://./conf/personal-merchant.yaml", "config uri. see: config.Server")
	}

}

// WithAppIdChannelConfigServer 设置AppIdChannelConfigServer
func WithAppIdChannelConfigServer(required bool) Option {
	return func(o *options) {
		if !required {
			return
		}
		u := o.AppIdChannelConfigServerURL
		s := string(o.AppIdChannelConfigServerURL)
		o.ps.StringVar(&s, u.Flag(), "file://./conf/app-id-channel-config.yaml", "config uri. see: config.Server")
	}
}
