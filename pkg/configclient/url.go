package configclient

type (
	// PayConfigServerURL PayConfigServerURL
	PayConfigServerURL string

	// NoticeConfigServerURL NoticeConfigServerURL
	NoticeConfigServerURL string

	// ServiceConfigServerURL ServiceConfigServerURL
	ServiceConfigServerURL string

	// ChannelServiceConfigServerURL ChannelServiceConfigServerURL
	ChannelServiceConfigServerURL string

	// MerchantConfigServerURL MerchantConfigServerURL
	MerchantConfigServerURL string

	// PersonalMerchantServerURL PersonalMerchantServerURL
	PersonalMerchantServerURL string

	// AppIdChannelConfigServerURL AppIdChannelConfigServerURL
	AppIdChannelConfigServerURL string
)

// ConfigURL 配置URL
type ConfigURL interface {
	URL() string
	Flag() string
}

func (u PayConfigServerURL) URL() string {
	return string(u)
}

func (u PayConfigServerURL) Flag() string {
	return "pay-config-url"
}

func (u NoticeConfigServerURL) URL() string {
	return string(u)
}

func (u NoticeConfigServerURL) Flag() string {
	return "notice-config-url"
}

func (u ServiceConfigServerURL) URL() string {
	return string(u)
}

func (u ServiceConfigServerURL) Flag() string {
	return "service-config-url"
}

func (u ChannelServiceConfigServerURL) URL() string {
	return string(u)
}

func (u ChannelServiceConfigServerURL) Flag() string {
	return "channel-service-config-url"
}

func (u MerchantConfigServerURL) URL() string {
	return string(u)
}

func (u MerchantConfigServerURL) Flag() string {
	return "merchant-config-url"
}

func (u PersonalMerchantServerURL) URL() string {
	return string(u)
}

func (u PersonalMerchantServerURL) Flag() string {
	return "personal-merchant-url"
}

func (u AppIdChannelConfigServerURL) URL() string {
	return string(u)
}

func (u AppIdChannelConfigServerURL) Flag() string {
	return "app-id-channel-config-url"
}
