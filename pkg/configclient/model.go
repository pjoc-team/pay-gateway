package configclient

// PayConfig 支付配置
type PayConfig struct {
	ClusterID        string `json:"cluster_id"`
	Concurrency      int    `json:"concurrency"`
	NotifyUrlPattern string `json:"notify_url_pattern"` // 通知地址的正则，必须包含{gateway_order_id}
	ReturnUrlPattern string `json:"return_url_pattern"` // 跳转地址的正则，必须包含{gateway_order_id}
}

// NoticeConfig 通知配置
type NoticeConfig struct {
	NoticeIntervalSecond int `json:"notice_interval_second"`
	// 通知间隔
	//
	// 例如: [30, 30, 120, 240, 480, 1200, 3600, 7200, 43200, 86400, 172800]
	// 表示如果通知失败，则会隔 30s, 30s, 2min, 4min, 8min, 20min, 1H, 2H, 12H, 24H, 48H 通知
	NoticeDelaySecondExpressions []int `json:"notice_expressions"`
}

// ServiceConfig 服务发现配置
type ServiceConfig struct {
	ServiceName string `json:"service_name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
}

// ChannelServiceConfig 渠道微服务配置
type ChannelServiceConfig struct {
	ChannelID   string `json:"channel_id"`
	ServiceName string `json:"service_name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
}

// MerchantConfig 商户配置
type MerchantConfig struct {
	AppID                string `json:"app_id"`
	GatewayRSAPublicKey  string `json:"gateway_rsa_public_key"`
	GatewayRSAPrivateKey string `json:"gateway_rsa_private_key"`
	MerchantRSAPublicKey string `json:"merchant_rsa_public_key"`
	Md5Key               string `json:"md5_key"`
}

// PersonalMerchant 个人码渠道
type PersonalMerchant struct {
	AppID                string `json:"app_id"`
	GatewayRSAPublicKey  string `json:"gateway_rsa_public_key"`
	GatewayRSAPrivateKey string `json:"gateway_rsa_private_key"`
	MerchantRSAPublicKey string `json:"merchant_rsa_public_key"`
	Md5Key               string `json:"md5_key"`
}

// AppIDChannelConfig AppID和渠道配置
type AppIDChannelConfig struct {
	RatePercent    float32 `json:"rate_percent"`
	Method         string  `json:"method"`
	ChannelAccount string  `json:"channel_account"`
	Available      bool    `json:"available"`
	ChannelID      string  `json:"channel_id"`
}
