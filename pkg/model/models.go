package model

import "github.com/pjoc-team/pay-gateway/pkg/configclient"

// ConfigType config type
type ConfigType string

// GatewayConfig gateway config
type GatewayConfig struct {
	// 集群配置
	PayConfig *configclient.PayConfig
	// 通知配置
	NoticeConfig *configclient.NoticeConfig
	// AppID和费率配置
	AppIDAndChannelConfigMap *AppIDAndChannelConfigMap
	// AppID和商户配置
	AppIDAndMerchantMap *MerchantConfigMap
	// 服务和对应的部署服务名映射
	ServiceMap *ServiceConfigMap
	// Channel和对应host配置
	ChannelServiceMap *ChannelServiceConfigMap
}

// FullPath full path of config
func (b *BaseConfig) FullPath() string {
	return b.baseDir + b.dirName
}

// Key key of config
func (b *BaseConfig) Key() string {
	return b.key
}

// BaseDir base dir of config
func (b *BaseConfig) BaseDir() string {
	return b.baseDir
}

// BaseConfig base config
type BaseConfig struct {
	key      string
	baseDir  string
	dirName  string
}

// ServiceConfigMap services
type ServiceConfigMap map[string]configclient.ServiceConfig

// ChannelServiceConfigMap channel services
type ChannelServiceConfigMap map[string]configclient.ChannelServiceConfig


// MerchantConfigMap merchant config
type MerchantConfigMap map[string]configclient.MerchantConfig

// PersonalMerchantConfigMap personal merchant config
type PersonalMerchantConfigMap map[string]configclient.PersonalMerchant
