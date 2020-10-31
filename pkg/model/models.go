package model

import "github.com/pjoc-team/pay-gateway/pkg/configclient"

type ConfigType string

type GatewayConfig struct {
	// 集群配置
	PayConfig *configclient.PayConfig
	// 通知配置
	NoticeConfig *configclient.NoticeConfig
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
	key      string
	baseDir  string
	dirName  string
}

// ################## service ##################
type ServiceConfigMap map[string]configclient.ServiceConfig

// ################## channel ##################
type ChannelServiceConfigMap map[string]configclient.ChannelServiceConfig

// ################## merchant ##################

type MerchantConfigMap map[string]configclient.MerchantConfig

// ################## personal merchant ##################
type PersonalMerchantConfigMap map[string]configclient.PersonalMerchant
