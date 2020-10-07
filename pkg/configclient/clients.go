package configclient

import (
	"github.com/pjoc-team/pay-gateway/pkg/model"
)

type ConfigClients struct {
}

func (c *ConfigClients) GetAppChannelConfig(appId string, method string) (*model.AppIdChannelConfig, error) {
	return nil, nil
}

func (c *ConfigClients) GetAppConfig(appId string) (*model.MerchantConfig, error) {
	return nil, nil
}
