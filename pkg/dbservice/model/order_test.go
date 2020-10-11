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
	orderRequest.BasePayOrder.GatewayOrderId = "123"
	order := &PayOrder{}
	copier.Copy(order, orderRequest)
	copier.Copy(order, orderRequest)
	fmt.Println(order.GatewayOrderId)
	assert.Equal(t, orderRequest.BasePayOrder.GatewayOrderId, order.GatewayOrderId)
}
