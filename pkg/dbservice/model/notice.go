package model

// Notice notice model
type Notice struct {
	GatewayOrderID string `gorm:"primary_key"`
	CreateDate     string `gorm:"index"`
	FailTimes      uint32
	NoticeTime     string `gorm:"index"`
	Status         string `gorm:"index"`
	ErrorMessage   string
	NextNotifyTime string `gorm:"index"`
}

// TableName table name of notice
func (Notice) TableName() string{
	return "pay_notice"
}

// NoticeOk notice ok model
type NoticeOk struct {
	GatewayOrderID string `gorm:"primary_key"`
	CreateDate     string `gorm:"index"`
	FailTimes      uint32
	NoticeTime     string `gorm:"index"`
}

// TableName table name of NoticeOk model
func (NoticeOk) TableName() string{
	return "pay_notice_ok"
}