package dbservice

import (
	"github.com/pjoc-team/pay-gateway/pkg/dbservice/model"
	pb "github.com/pjoc-team/pay-proto/go"
)

func newDbNotice(payNotice *pb.PayNotice) *model.Notice {
	notice := &model.Notice{
		GatewayOrderID: payNotice.GatewayOrderId,
		CreateDate:     payNotice.CreateDate,
		FailTimes:      payNotice.FailTimes,
		NoticeTime:     payNotice.NoticeTime,
		Status:         payNotice.Status,
		ErrorMessage:   payNotice.ErrorMessage,
		NextNotifyTime: payNotice.NextNotifyTime,
	}
	return notice
}

func newDbNoticeOk(payNoticeOk *pb.PayNoticeOk) *model.NoticeOk {
	noticeOk := &model.NoticeOk{
		GatewayOrderID: payNoticeOk.GatewayOrderId,
		CreateDate:     payNoticeOk.CreateDate,
		FailTimes:      payNoticeOk.FailTimes,
		NoticeTime:     payNoticeOk.NoticeTime,
	}
	return noticeOk
}

func newDbPayOrder(payOrder *pb.PayOrder) *model.PayOrder {
	order := &model.PayOrder{
		BasePayOrder: model.BasePayOrder{
			Version:             payOrder.BasePayOrder.Version,
			OutTradeNo:          payOrder.BasePayOrder.OutTradeNo,
			ChannelAccount:      payOrder.BasePayOrder.ChannelAccount,
			ChannelOrderID:      payOrder.BasePayOrder.ChannelOrderId,
			GatewayOrderID:      payOrder.BasePayOrder.GatewayOrderId,
			PayAmount:           payOrder.BasePayOrder.PayAmount,
			Currency:            payOrder.BasePayOrder.Currency,
			NotifyURL:           payOrder.BasePayOrder.NotifyUrl,
			ReturnURL:           payOrder.BasePayOrder.ReturnUrl,
			AppID:               payOrder.BasePayOrder.AppId,
			SignType:            payOrder.BasePayOrder.SignType,
			OrderTime:           payOrder.BasePayOrder.OrderTime,
			RequestTime:         payOrder.BasePayOrder.RequestTime,
			CreateDate:          payOrder.BasePayOrder.CreateDate,
			UserIP:              payOrder.BasePayOrder.UserIp,
			UserID:              payOrder.BasePayOrder.UserId,
			PayerAccount:        payOrder.BasePayOrder.PayerAccount,
			ProductID:           payOrder.BasePayOrder.ProductId,
			ProductName:         payOrder.BasePayOrder.ProductName,
			ProductDescribe:     payOrder.BasePayOrder.ProductDescribe,
			CallbackJSON:        payOrder.BasePayOrder.CallbackJson,
			ExtJSON:             payOrder.BasePayOrder.ExtJson,
			ChannelResponseJSON: payOrder.BasePayOrder.ChannelResponseJson,
			ErrorMessage:        payOrder.BasePayOrder.ErrorMessage,
			ChannelID:           payOrder.BasePayOrder.ChannelId,
			Method:              payOrder.BasePayOrder.Method,
			Remark:              payOrder.BasePayOrder.Remark,
		},
		OrderStatus: payOrder.OrderStatus,
	}
	return order
}

func newDbPayOrderOk(payOrderOk *pb.PayOrderOk) *model.PayOrderOk {
	orderOk := &model.PayOrderOk{
		BasePayOrder: model.BasePayOrder{
			Version:             payOrderOk.BasePayOrder.Version,
			OutTradeNo:          payOrderOk.BasePayOrder.OutTradeNo,
			ChannelAccount:      payOrderOk.BasePayOrder.ChannelAccount,
			ChannelOrderID:      payOrderOk.BasePayOrder.ChannelOrderId,
			GatewayOrderID:      payOrderOk.BasePayOrder.GatewayOrderId,
			PayAmount:           payOrderOk.BasePayOrder.PayAmount,
			Currency:            payOrderOk.BasePayOrder.Currency,
			NotifyURL:           payOrderOk.BasePayOrder.NotifyUrl,
			ReturnURL:           payOrderOk.BasePayOrder.ReturnUrl,
			AppID:               payOrderOk.BasePayOrder.AppId,
			SignType:            payOrderOk.BasePayOrder.SignType,
			OrderTime:           payOrderOk.BasePayOrder.OrderTime,
			RequestTime:         payOrderOk.BasePayOrder.RequestTime,
			CreateDate:          payOrderOk.BasePayOrder.CreateDate,
			UserIP:              payOrderOk.BasePayOrder.UserIp,
			UserID:              payOrderOk.BasePayOrder.UserId,
			PayerAccount:        payOrderOk.BasePayOrder.PayerAccount,
			ProductID:           payOrderOk.BasePayOrder.ProductId,
			ProductName:         payOrderOk.BasePayOrder.ProductName,
			ProductDescribe:     payOrderOk.BasePayOrder.ProductDescribe,
			CallbackJSON:        payOrderOk.BasePayOrder.CallbackJson,
			ExtJSON:             payOrderOk.BasePayOrder.ExtJson,
			ChannelResponseJSON: payOrderOk.BasePayOrder.ChannelResponseJson,
			ErrorMessage:        payOrderOk.BasePayOrder.ErrorMessage,
			ChannelID:           payOrderOk.BasePayOrder.ChannelId,
			Method:              payOrderOk.BasePayOrder.Method,
			Remark:              payOrderOk.BasePayOrder.Remark,
		},
		SuccessTime:     payOrderOk.SuccessTime,
		BalanceDate:     payOrderOk.BalanceDate,
		FareAmt:         payOrderOk.FareAmt,
		FactAmt:         payOrderOk.FactAmt,
		SendNoticeStats: payOrderOk.SendNoticeStats,
	}
	return orderOk
}
