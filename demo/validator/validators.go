package main

import (
	"fmt"
	"github.com/pjoc-team/pay-gateway/pkg/validator"
	"reflect"
)

func main() {
	for _, v := range validator.Validators {
		fmt.Println(reflect.TypeOf(v))
	}
}
