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
	Required() bool
}

// configURL 配置说明
type configURL struct {
	url      string
	required bool
	flag     string
}

func (c configURL) URL() string {
	return c.url
}

func (c configURL) Flag() string {
	return c.flag
}

func (c configURL) Required() bool {
	return c.required
}
