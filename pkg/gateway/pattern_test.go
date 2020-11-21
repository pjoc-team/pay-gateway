package gateway

import (
	"github.com/blademainer/commons/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReplaceGatewayOrderID(t *testing.T) {
	gatewayOrderID := util.RandString(64)
	id := ReplaceGatewayOrderID("http://127.0.0.1:8888/notify/{gateway_order_id}", gatewayOrderID)
	assert.Equal(t, "http://127.0.0.1:8888/notify/"+gatewayOrderID, id)
}

func BenchmarkReplaceGatewayOrderID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gatewayOrderID := util.RandString(64)
		ReplaceGatewayOrderID("http://127.0.0.1:8888/notify/{gateway_order_id}", gatewayOrderID)
	}
}

func BenchmarkReplacePlaceholder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gatewayOrderID := util.RandString(64)
		_, err := ReplacePlaceholder("http://127.0.0.1:8888/notify/{gateway_order_id}",
			"gateway_order_id",
			gatewayOrderID)
		if err != nil{
			b.Fatal(err.Error())
		}
	}
}
