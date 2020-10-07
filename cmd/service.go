package wired

import (
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	pay "github.com/pjoc-team/pay-proto/go"
)

// InitConfigClients 初始化配置
func InitConfigClients() (*configclient.ConfigClients, error) {
	return &configclient.ConfigClients{}, nil
}

func InitDbClient() (pay.PayDatabaseServiceClient, error) {
	return nil, nil
}
