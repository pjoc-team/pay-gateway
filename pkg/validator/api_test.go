package validator

import (
	"fmt"
	"testing"
)

func TestGetValidators(t *testing.T) {
	for _, v := range Validators {
		fmt.Println(v)
	}
}
