package notify

import (
	"context"
	"github.com/go-playground/form"
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-gateway/pkg/sign"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"net/url"
)

// Body notify body
type Body struct {
	Version string `json:"version,omitempty" form:"version"`
	// 业务订单号
	OutTradeNo string `json:"out_trade_no,omitempty" form:"out_trade_no"`
	// 支付金额（分）
	PayAmount uint32 `json:"pay_amount,omitempty" form:"pay_amount"`
	// 币种
	Currency string `json:"currency,omitempty" form:"currency"`
	// 支付后跳转的前端地址
	ReturnUrl string `json:"return_url,omitempty" form:"return_url"`
	// 系统给商户分配的app_id
	AppId string `json:"app_id,omitempty" form:"app_id"`
	// 加密方法，RSA和MD5，默认RSA
	SignType string `json:"sign_type,omitempty" form:"sign_type"`
	// 签名
	Sign string `json:"sign,omitempty" form:"sign"`
	// 业务方下单时间，时间格式: 年年年年-月月-日日 时时:分分:秒秒，例如: 2006-01-02 15:04:05
	OrderTime string `json:"order_time,omitempty" form:"order_time"`
	// 发起支付的用户ip
	UserIp string `json:"user_ip,omitempty" form:"user_ip"`
	// 用户在业务系统的id
	UserId string `json:"user_id,omitempty" form:"user_id"`
	// 支付者账号，可选
	PayerAccount string `json:"payer_account,omitempty" form:"payer_account"`
	// 业务系统的产品id
	ProductId string `json:"product_id,omitempty" form:"product_id"`
	// 商品名称
	ProductName string `json:"product_name,omitempty" form:"product_name"`
	// 商品描述
	ProductDescribe string `json:"product_describe,omitempty" form:"product_describe"`
	// 参数编码，只允许utf-8编码；签名时一定要使用该编码获取字节然后再进行签名
	Charset string `json:"charset,omitempty" form:"charset"`
	// 回调业务系统时需要带上的字符串
	CallbackJson string `json:"callback_json,omitempty" form:"callback_json"`
	// 扩展json
	ExtJson string `json:"ext_json,omitempty" form:"ext_json"`
	// 渠道id（非必须），如果未指定method，系统会根据method来找到可用的channel_id
	ChannelId string `json:"channel_id,omitempty" form:"channel_id"`
	// 例如：二维码支付，银联支付等。
	Method string `json:"method,omitempty" form:"method"`
	// 实际金额
	FactAmt uint32 `json:"fact_amt,omitempty" form:"fact_amt"`
	// 手续费
	FareAmt uint32 `json:"fare_amt,omitempty" form:"fare_amt"`
}

// UrlGenerator generate notify url
type UrlGenerator struct {
	// validator.paramsCompacter = sign.NewParamsCompacter(&pay.PayRequest{}, "json", []string{"sign"}, true, "&", "=")
	paramsCompacter sign.ParamsCompacter
	config          configclient.ConfigClients
	encoder         *form.Encoder
}

// NewUrlGenerator create url generator
func NewUrlGenerator(config configclient.ConfigClients) *UrlGenerator {
	generator := &UrlGenerator{}
	generator.paramsCompacter = sign.NewParamsCompacter(
		&Body{}, "json", []string{"sign"}, true, "&", "=",
	)
	generator.config = config
	generator.encoder = form.NewEncoder()
	return generator
}

// GenerateSign generate sign
func (g *UrlGenerator) GenerateSign(ctx context.Context, body *Body) (str string, err error) {
	log := logger.Log()

	appConfig, err := g.config.GetAppConfig(ctx, body.AppId)
	if err != nil {
		log.Errorf("failed to get config: %v error: %v", body.AppId, err.Error())
		return "", err
	}

	source := g.paramsCompacter.ParamsToString(body)
	log.Debugf("Generate url string: %v by body: %v", source, body)
	str, err = sign.GenerateSign(ctx, body.Charset, source, appConfig, sign.Type(body.SignType))
	return
}

// GenerateSignByPayOrderOk sign by ok order
func (g *UrlGenerator) GenerateSignByPayOrderOk(
	ctx context.Context, payOrderOk pay.PayOrderOk,
) (str string, err error) {
	log := logger.ContextLog(ctx)

	body := &Body{}
	if err = copier.Copy(body, payOrderOk); err != nil {
		log.Errorf("Failed to copy order to body! order: %v error: %v", payOrderOk, err.Error())
		return
	}
	return g.GenerateSign(ctx, body)
}

// GenerateUrlByPayOrderOk generate url by ok order
func (g *UrlGenerator) GenerateUrlByPayOrderOk(
	ctx context.Context, payOrderOk pay.PayOrderOk,
) (url string, form url.Values, err error) {
	log := logger.ContextLog(ctx)

	url = payOrderOk.BasePayOrder.NotifyUrl
	body := &Body{}
	if err = copier.Copy(body, payOrderOk.BasePayOrder); err != nil {
		log.Errorf("Failed to copy order to body! order: %v error: %v", payOrderOk, err.Error())
		return
	}
	signMessage, err := g.GenerateSign(ctx, body)
	if err != nil {
		return
	}
	body.Sign = signMessage
	form, err = g.encoder.Encode(body)
	if err != nil {
		log.Errorf("Failed to generate form! body: %v error: %v", body, err.Error())
	} else {
		log.Infof("Generate form: %v by order: %v", form, payOrderOk)
	}
	return
}
