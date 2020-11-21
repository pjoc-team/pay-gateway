package model

import "github.com/pjoc-team/pay-gateway/pkg/configclient"

// AppIDAndChannelConfigMap app id and channel configs
type AppIDAndChannelConfigMap map[string]AppIDAndChannelConfigs

// AppIDAndChannelConfigs app id and channel config
type AppIDAndChannelConfigs struct {
	AppID          string                            `json:"app_id"`
	ChannelConfigs []configclient.AppIDChannelConfig `json:"channel_configs"`
}
