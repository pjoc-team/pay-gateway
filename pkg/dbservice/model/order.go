package model

// BasePayOrder base order info
type BasePayOrder struct {
	Version string `gorm:"-" json:"version"`
	// 业务订单号
	OutTradeNo string `gorm:"unique_index:idx_app_id_out_trade_no" json:"out_trade_no"`
	// 渠道账号
	ChannelAccount string `json:"channel_account"`
	// 渠道订单号
	ChannelOrderID string `json:"channel_order_id"`
	// 网关订单号
	GatewayOrderID string `gorm:"primary_key" json:"gateway_order_id"`
	// 支付金额（分）
	PayAmount uint32 `json:"pay_amount"`
	// 币种
	Currency string `json:"currency"`
	// 接收通知的地址，不能带参数（即：不能包含问号）
	NotifyURL string `json:"notify_url"`
	// 支付后跳转的前端地址
	ReturnURL string `json:"return_url"`
	// 系统给商户分配的app_id
	AppID string `gorm:"unique_index:idx_app_id_out_trade_no" json:"app_id"`
	// 加密方法，rsa和md5，默认rsa
	SignType string `json:"sign_type"`
	// 下单时间
	OrderTime string `json:"order_time"`
	// 请求到网关的时间
	RequestTime string `json:"request_time"`
	// 订单创建日期
	CreateDate string `gorm:"index" json:"create_date"`
	// 发起支付的用户ip
	UserIP string `json:"user_ip"`
	// 用户在业务系统的id
	UserID string `gorm:"index" json:"user_id"`
	// 支付者账号，可选
	PayerAccount string `json:"payer_account"`
	// 产品id
	ProductID string `json:"product_id"`
	// 商品名称
	ProductName string `json:"product_name"`
	// 商品描述
	ProductDescribe string `json:"product_describe"`
	// 回调业务系统时需要带上的字符串
	CallbackJSON string `sql:"type:text;" json:"callback_json"`
	// 扩展json
	ExtJSON string `sql:"type:text;" json:"ext_json"`
	// 渠道返回的json
	ChannelResponseJSON string `sql:"type:text;" json:"channel_response_json"`
	// 下单错误信息
	ErrorMessage string `sql:"type:text;" json:"error_message"`
	// 渠道id（非必须），如果未指定method，系统会根据method来找到可用的channel_id
	ChannelID string `gorm:"index" json:"channel_id"`
	Method    string `gorm:"index" json:"method"`
	// 备注
	Remark    string `sql:"type:text;" json:"remark"`
}

// PayOrder pay order model
type PayOrder struct {
	BasePayOrder
	OrderStatus string `gorm:"index" json:"order_status"`
}

// TableName table name of pay_order
func (PayOrder) TableName() string{
	return "pay_order"
}

// PayOrderOk pay order ok model
type PayOrderOk struct {
	BasePayOrder
	SuccessTime     string `json:"success_time"`
	BalanceDate     string `gorm:"index" json:"balance_date"`
	FareAmt         uint32 `json:"fare_amt"`
	FactAmt         uint32 `json:"fact_amt"`
	SendNoticeStats string `gorm:"index" json:"send_notice_stats"`
}

// TableName table name of PayOrderOk
func (PayOrderOk) TableName() string{
	return "pay_order_ok"
}
