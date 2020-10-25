package model

import "github.com/pjoc-team/pay-gateway/pkg/configclient"

// ################## appid ##################
type AppIdAndChannelConfigMap map[string]AppIdAndChannelConfigs

type AppIdAndChannelConfigs struct {
	AppId          string                            `json:"app_id"`
	ChannelConfigs []configclient.AppIDChannelConfig `json:"channel_configs"`
}
