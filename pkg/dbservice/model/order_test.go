package model

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/pay-proto/go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCopy(t *testing.T) {
	orderRequest := &pay.PayOrder{}
	orderRequest.BasePayOrder = &pay.BasePayOrder{}
	orderRequest.BasePayOrder.GatewayOrderID = "123"
	order := &PayOrder{}
	err := copier.Copy(order, orderRequest)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = copier.Copy(order, orderRequest)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(order.GatewayOrderID)
	assert.Equal(t, orderRequest.BasePayOrder.GatewayOrderID, order.GatewayOrderID)
}
