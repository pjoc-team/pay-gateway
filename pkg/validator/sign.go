package validator

import "github.com/pjoc-team/base-service/pkg/sign"

func init() {
	InitCheckSignValidator()
}

func InitCheckSignValidator() {
	validator := sign.NewCheckSignValidator()
	RegisterValidator(validator)
}
