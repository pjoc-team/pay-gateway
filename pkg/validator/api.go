package validator

import (
	"context"
	"github.com/pjoc-team/pay-gateway/pkg/model"
	"github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
)

type GetMerchantConfigFunc func(appID string) (*model.MerchantConfig, error)

type Validator interface {
	Validate(ctx context.Context, request pay.PayRequest, cfg GetMerchantConfigFunc) error
}

var Validators = make([]Validator, 0)

func RegisterValidator(validator Validator) {
	logger.Log().Infof("register validator: %v", validator)
	Validators = append(Validators, validator)
}

func Validate(ctx context.Context, request pay.PayRequest, cfg GetMerchantConfigFunc) error {
	for _, validator := range Validators {
		if e := validator.Validate(ctx, request, cfg); e != nil {
			return e
		}
	}
	return nil
}
