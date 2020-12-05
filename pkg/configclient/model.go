package configclient

// PayConfig 支付配置
type PayConfig struct {
	ClusterID   string `json:"cluster_id" yaml:"clusterID"`
	Concurrency int    `json:"concurrency" yaml:"concurrency"`
	// 通知地址的正则，必须包含{gateway_order_id}
	NotifyURLPattern string `json:"notify_url_pattern" yaml:"notifyURLPattern" validate:"required"`
	// 跳转地址的正则，必须包含{gateway_order_id}
	ReturnURLPattern string `json:"return_url_pattern" yaml:"returnURLPattern" validate:"required"`
}

// NotifyConfig 通知配置
type NotifyConfig struct {
	NotifyIntervalSecond int `json:"notify_interval_second" yaml:"notifyIntervalSecond"`
	// 通知间隔
	//
	// 例如: [30, 30, 120, 240, 480, 1200, 3600, 7200, 43200, 86400, 172800]
	// 表示如果通知失败，则会隔 30s, 30s, 2min, 4min, 8min, 20min, 1H, 2H, 12H, 24H, 48H 通知
	NotifyDelaySecondExpressions []int `json:"notify_expressions" yaml:"notifyDelaySecondExpressions"`
}

// ServiceConfig 服务发现配置
type ServiceConfig struct {
	ServiceName string `json:"service_name" yaml:"serviceName"`
	Host        string `json:"host" yaml:"host"`
	Port        int    `json:"port" yaml:"port"`
}

// ChannelServiceConfig 渠道微服务配置
type ChannelServiceConfig struct {
	ChannelID   string `json:"channel_id" yaml:"channelID"  validate:"required"`
	ServiceName string `json:"service_name" yaml:"serviceName"  validate:"required"`
}

// MerchantConfig 商户配置
type MerchantConfig struct {
	AppID                string `json:"app_id" yaml:"appID"`
	GatewayRSAPublicKey  string `json:"gateway_rsa_public_key" yaml:"gatewayRsaPublicKey"`
	GatewayRSAPrivateKey string `json:"gateway_rsa_private_key" yaml:"gatewayRsaPrivateKey"`
	MerchantRSAPublicKey string `json:"merchant_rsa_public_key" yaml:"merchantRsaPublicKey"`
	Md5Key               string `json:"md5_key" yaml:"md5Key"`
}

// PersonalMerchant 个人码渠道
type PersonalMerchant struct {
	AppID                string `json:"app_id" yaml:"appID"`
	GatewayRSAPublicKey  string `json:"gateway_rsa_public_key" yaml:"gatewayRsaPublicKey"`
	GatewayRSAPrivateKey string `json:"gateway_rsa_private_key" yaml:"gatewayRsaPrivateKey"`
	MerchantRSAPublicKey string `json:"merchant_rsa_public_key" yaml:"merchantRsaPublicKey"`
	Md5Key               string `json:"md5_key" yaml:"md5Key"`
}

// AppIDChannelConfig AppID和渠道配置
type AppIDChannelConfig struct {
	RatePercent    float32 `json:"rate_percent" yaml:"ratePercent"`
	Method         string  `json:"method" yaml:"method"`
	ChannelAccount string  `json:"channel_account" yaml:"channelAccount"`
	Available      bool    `json:"available" yaml:"available"`
	ChannelID      string  `json:"channel_id" yaml:"channelID"`
}
