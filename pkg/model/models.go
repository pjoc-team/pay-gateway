package model

type ConfigType string

type GatewayConfig struct {
	// 集群配置
	PayConfig *PayConfig
	// 通知配置
	NoticeConfig *NoticeConfig
	// AppId和费率配置
	AppIdAndChannelConfigMap *AppIdAndChannelConfigMap
	// AppId和商户配置
	AppIdAndMerchantMap *MerchantConfigMap
	// 服务和对应的部署服务名映射
	ServiceMap *ServiceConfigMap
	// Channel和对应host配置
	ChannelServiceMap *ChannelServiceConfigMap
}

func (b *BaseConfig) FullPath() string {
	return b.baseDir + b.dirName
}

func (b *BaseConfig) Key() string {
	return b.key
}

func (b *BaseConfig) BaseDir() string {
	return b.baseDir
}

type BaseConfig struct {
	key      string `json:"-"`
	baseDir  string `json:"-"`
	dirName  string `json:"-"`
	fullPath string `json:"-"`
}

type PayConfig struct {
	ClusterId        string `json:"cluster_id"`
	Concurrency      int    `json:"concurrency"`
	NotifyUrlPattern string `json:"notify_url_pattern"` // 通知地址的正则，必须包含{gateway_order_id}
	ReturnUrlPattern string `json:"return_url_pattern"` // 跳转地址的正则，必须包含{gateway_order_id}
}

// ################## notice ##################
type NoticeConfig struct {
	NoticeIntervalSecond int `json:"notice_interval_second"`
	// 通知间隔
	//
	// 例如: [30, 30, 120, 240, 480, 1200, 3600, 7200, 43200, 86400, 172800]
	// 表示如果通知失败，则会隔 30s, 30s, 2min, 4min, 8min, 20min, 1H, 2H, 12H, 24H, 48H 通知
	NoticeDelaySecondExpressions []int `json:"notice_expressions"`
}

// ################## service ##################
type ServiceConfigMap map[string]ServiceConfig
type ServiceConfig struct {
	ServiceName string `json:"service_name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
}

// ################## channel ##################
type ChannelServiceConfigMap map[string]ChannelServiceConfig

type ChannelServiceConfig struct {
	BaseConfig
	ChannelId   string `json:"channel_id"`
	ServiceName string `json:"service_name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
}

// ################## merchant ##################

type MerchantConfigMap map[string]MerchantConfig

type MerchantConfig struct {
	AppId                string `json:"app_id"`
	GatewayRSAPublicKey  string `json:"gateway_rsa_public_key"`
	GatewayRSAPrivateKey string `json:"gateway_rsa_private_key"`
	MerchantRSAPublicKey string `json:"merchant_rsa_public_key"`
	Md5Key               string `json:"md5_key"`
}

// ################## personal merchant ##################
type PersonalMerchantConfigMap map[string]PersonalMerchant

type PersonalMerchant struct {
	AppId                string `json:"app_id"`
	GatewayRSAPublicKey  string `json:"gateway_rsa_public_key"`
	GatewayRSAPrivateKey string `json:"gateway_rsa_private_key"`
	MerchantRSAPublicKey string `json:"merchant_rsa_public_key"`
	Md5Key               string `json:"md5_key"`
}
