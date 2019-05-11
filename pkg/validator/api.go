package validator

import (
	"github.com/pjoc-team/base-service/pkg/logger"
	"github.com/pjoc-team/base-service/pkg/service"
	"github.com/pjoc-team/pay-proto/go"
)

type Validator interface {
	Validate(request pay.PayRequest, cfg service.GatewayConfig) error
}

var Validators = make([]Validator, 0)

func RegisterValidator(validator Validator) {
	logger.Log.Infof("register validator: %v", validator)
	Validators = append(Validators, validator)
}

func Validate(request pay.PayRequest, cfg service.GatewayConfig) error {
	for _, validator := range Validators {
		if e := validator.Validate(request, cfg); e != nil {
			return e
		}
	}
	return nil
}
