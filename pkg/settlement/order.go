package settlement

import (
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	pb "github.com/pjoc-team/pay-proto/go"
	"math"
	"math/big"
)

func (svc *service) GenerateSuccessOrder(order *pb.PayOrder) *pb.PayOrderOk {
	orderOk := &pb.PayOrderOk{}
	copier.Copy(orderOk, order)
	if config := svc.findMerchantConfig(order); config == nil {
		rate := config.RatePercent
		orderOk.FactAmt, orderOk.FareAmt = calculateFactAmt(order.BasePayOrder.PayAmount, rate)
	}
	return orderOk
}

func (svc *service) findMerchantConfig(order *pb.PayOrder) *configclient.AppIDChannelConfig {
	configMap := *svc.AppIdAndChannelConfigMap
	if configMap == nil {
		return nil
	}
	merchantConfig := configMap[order.BasePayOrder.AppId]
	for _, channelConfig := range merchantConfig.ChannelConfigs {
		if channelConfig.ChannelId == order.BasePayOrder.ChannelId {
			return &channelConfig
		}
	}
	return nil
}

func calculateFactAmt(orderAmt uint32, ratePercent float32) (factAmt uint32, fareAmt uint32) {
	rateFloat := big.NewFloat(float64(ratePercent))
	orderAmtFloat := big.NewFloat(float64(orderAmt))
	fareAmtFloat := big.NewFloat(0).Mul(rateFloat, orderAmtFloat)
	fareAmtFloat = big.NewFloat(0).Quo(fareAmtFloat, big.NewFloat(100))
	factAmtFloat := big.NewFloat(0).Add(
		orderAmtFloat, big.NewFloat(0).Mul(fareAmtFloat, big.NewFloat(-1)),
	)

	factAmt64, _ := factAmtFloat.Float64()
	factAmt = round(factAmt64)

	fareAmt64, _ := fareAmtFloat.Float64()
	fareAmt = round(fareAmt64)
	return
}

func round(f float64) uint32 {
	floor := math.Floor(f + 0.5)
	return uint32(floor)
}
