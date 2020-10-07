package gateway

type AppIdChannelConfig struct {
	RatePercent    float32 `json:"rate_percent"`
	Method         string  `json:"method"`
	ChannelAccount string  `json:"channel_account"`
	Available      bool    `json:"available"`
	ChannelId      string  `json:"channel_id"`
}


type AppIdAndChannelConfigMap map[string]AppIdAndChannelConfigs

type AppIdAndChannelConfigs struct {
	AppId          string               `json:"app_id"`
	ChannelConfigs []AppIdChannelConfig `json:"channel_configs"`
}

type ChannelServiceConfigMap map[string]ChannelServiceConfig

type ChannelServiceConfig struct {
	ChannelId   string `json:"channel_id"`
	ServiceName string `json:"service_name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
}