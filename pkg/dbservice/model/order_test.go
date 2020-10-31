package model

//func TestCopy(t *testing.T) {
//	orderRequest := &pay.PayOrder{}
//	orderRequest.BasePayOrder = &pay.BasePayOrder{}
//	orderRequest.BasePayOrder.GatewayOrderId = "123"
//	order := &PayOrder{}
//	err := copier.Copy(order, orderRequest)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//	err = copier.Copy(order, orderRequest)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//	fmt.Println(order.GatewayOrderID)
//	assert.Equal(t, orderRequest.BasePayOrder.GatewayOrderId, order.GatewayOrderID)
//}
