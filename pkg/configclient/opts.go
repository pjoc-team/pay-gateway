package configclient

import "github.com/spf13/pflag"

// options 配置选项
type options struct {
	ps                            *pflag.FlagSet
	PayConfigServerURL            *configURL
	NoticeConfigServerURL         *configURL
	ServiceConfigServerURL        *configURL
	ChannelServiceConfigServerURL *configURL
	MerchantConfigServerURL       *configURL
	PersonalMerchantServerURL     *configURL
	AppIdChannelConfigServerURL   *configURL
}

func newOpts() (*options, error) {
	set := pflag.NewFlagSet("config-url", pflag.PanicOnError)

	o := &options{
		ps: set,
	}

	o.PayConfigServerURL = &configURL{
		required: false,
		flag:     "pay-config-url",
	}
	o.ps.StringVar(&o.PayConfigServerURL.url, o.PayConfigServerURL.Flag(), "file://./conf/biz/pay-config.yaml", "config uri. see: config.Server")

	o.NoticeConfigServerURL = &configURL{
		required: false,
		flag:     "notice-config-url",
	}
	o.ps.StringVar(&o.NoticeConfigServerURL.url, o.NoticeConfigServerURL.Flag(), "file://./conf/biz/notice-config.yaml", "config uri. see: config.Server")

	o.ServiceConfigServerURL = &configURL{
		required: false,
		flag:     "service-config-url",
	}
	o.ps.StringVar(&o.ServiceConfigServerURL.url, o.ServiceConfigServerURL.Flag(), "file://./conf/biz/service-config.yaml", "config uri. see: config.Server")

	o.ChannelServiceConfigServerURL = &configURL{
		required: false,
		flag:     "channel-service-config-url",
	}
	o.ps.StringVar(&o.ChannelServiceConfigServerURL.url, o.ChannelServiceConfigServerURL.Flag(), "file://./conf/biz/channel-service-config.yaml", "config uri. see: config.Server")

	o.MerchantConfigServerURL = &configURL{
		required: false,
		flag:     "merchant-config-url",
	}
	o.ps.StringVar(&o.MerchantConfigServerURL.url, o.MerchantConfigServerURL.Flag(), "file://./conf/biz/merchant-config.yaml", "config uri. see: config.Server")

	o.PersonalMerchantServerURL = &configURL{
		required: false,
		flag:     "personal-merchant-url",
	}
	o.ps.StringVar(&o.PersonalMerchantServerURL.url, o.PersonalMerchantServerURL.Flag(), "file://./conf/biz/personal-merchant.yaml", "config uri. see: config.Server")

	o.AppIdChannelConfigServerURL = &configURL{
		required: false,
		flag:     "app-id-channel-config-url",
	}
	o.ps.StringVar(&o.AppIdChannelConfigServerURL.url, o.AppIdChannelConfigServerURL.Flag(), "file://./conf/biz/app-id-channel-config.yaml", "config uri. see: config.Server")

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
		o.PayConfigServerURL.required = required
	}
}

// WithNoticeConfigServer 设置NoticeConfigServer
func WithNoticeConfigServer(required bool) Option {
	return func(o *options) {
		o.NoticeConfigServerURL.required = required
	}
}

// WithServiceConfigServer 设置ServiceConfigServer
func WithServiceConfigServer(required bool) Option {
	return func(o *options) {
		o.ServiceConfigServerURL.required = required
	}
}

// WithChannelServiceConfigServer 设置ChannelServiceConfigServer
func WithChannelServiceConfigServer(required bool) Option {
	return func(o *options) {
		o.ChannelServiceConfigServerURL.required = required
	}
}

// WithMerchantConfigServer 设置MerchantConfigServer
func WithMerchantConfigServer(required bool) Option {
	return func(o *options) {
		o.MerchantConfigServerURL.required = required
	}
}

// WithPersonalMerchantServer 设置PersonalMerchantServer
func WithPersonalMerchantServer(required bool) Option {
	return func(o *options) {
		o.PersonalMerchantServerURL.required = required
	}

}

// WithAppIdChannelConfigServer 设置AppIdChannelConfigServer
func WithAppIdChannelConfigServer(required bool) Option {
	return func(o *options) {
		o.AppIdChannelConfigServerURL.required = required
	}
}
