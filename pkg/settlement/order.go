package settlement

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"math"
	"math/big"
)

// GenerateSuccessOrder generate success order of pay order
func (svc *service) GenerateSuccessOrder(ctx context.Context, order *pb.PayOrder) (
	*pb.PayOrderOk, error,
) {
	log := logger.ContextLog(ctx)
	orderOk := &pb.PayOrderOk{}
	err2 := copier.Copy(orderOk, order)
	if err2 != nil {
		panic(err2.Error())
	}
	config, err := svc.findMerchantConfig(ctx, order)
	if err != nil {
		log.Errorf("not found merchant config of order: %v error: %v", order, err.Error())
		return nil, err
	} else if config == nil {
		err = fmt.Errorf("not found merchant config")
		return nil, err
	}
	rate := config.RatePercent
	orderOk.FactAmt, orderOk.FareAmt = calculateFactAmt(order.BasePayOrder.PayAmount, rate)
	return orderOk, nil
}

func (svc *service) findMerchantConfig(
	ctx context.Context, order *pb.PayOrder,
) (*configclient.AppIDChannelConfig, error) {
	log := logger.ContextLog(ctx)
	channelConfigs, err := svc.config.GetAppChannelConfig(
		ctx, order.BasePayOrder.AppId,
		order.BasePayOrder.Method,
	)
	if err != nil {
		log.Errorf(
			"failed to get appChannelConfig, appID: %v method: %v error: %v",
			order.BasePayOrder.AppId, order.BasePayOrder.Method, err.Error(),
		)
		return nil, err
	}
	for _, channelConfig := range channelConfigs {
		if channelConfig.ChannelID == order.BasePayOrder.ChannelId {
			return channelConfig, nil
		}
	}
	return nil, nil
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
