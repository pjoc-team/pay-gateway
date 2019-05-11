package validator

import (
	"github.com/pjoc-team/base-service/pkg/service"
	"github.com/pjoc-team/pay-proto/go"
)

func init() {
	RegisterValidator(&ParamsValidator{})
}

type ParamsValidator struct {
}

func (p *ParamsValidator) Validate(request pay.PayRequest, cfg service.GatewayConfig) (err error) {
	return
}
