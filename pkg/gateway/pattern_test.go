package gateway

import (
	"github.com/blademainer/commons/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReplaceGatewayOrderId(t *testing.T) {
	gatewayOrderId := util.RandString(64)
	id := ReplaceGatewayOrderId("http://127.0.0.1:8888/notify/{gateway_order_id}", gatewayOrderId)
	assert.Equal(t, "http://127.0.0.1:8888/notify/"+gatewayOrderId, id)
}

func BenchmarkReplaceGatewayOrderId(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gatewayOrderId := util.RandString(64)
		ReplaceGatewayOrderId("http://127.0.0.1:8888/notify/{gateway_order_id}", gatewayOrderId)
	}
}

func BenchmarkReplacePlaceholder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gatewayOrderId := util.RandString(64)
		ReplacePlaceholder("http://127.0.0.1:8888/notify/{gateway_order_id}", "gateway_order_id", gatewayOrderId)
	}
}
