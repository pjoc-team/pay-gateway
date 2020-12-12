package validator

import (
	"context"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
)

// GetMerchantConfigFunc func to get merchant
type GetMerchantConfigFunc func(ctx context.Context, appID string) (*configclient.MerchantConfig, error)

// Validator validator interface
type Validator interface {
	Validate(ctx context.Context, request *pay.PayRequest, cfg GetMerchantConfigFunc) error
}

// Validators validators
var Validators = make([]Validator, 0)

// RegisterValidator register validator
func RegisterValidator(validator Validator) {
	logger.Log().Infof("register validator: %v", validator)
	Validators = append(Validators, validator)
}

// Validate validate pay request
func Validate(ctx context.Context, request *pay.PayRequest, cfg func(ctx context.Context,
	appID string) (*configclient.MerchantConfig, error)) error {
	for _, validator := range Validators {
		if e := validator.Validate(ctx, request, cfg); e != nil {
			return e
		}
	}
	return nil
}
