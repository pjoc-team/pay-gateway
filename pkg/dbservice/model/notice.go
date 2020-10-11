package model

type Notice struct {
	GatewayOrderId string `gorm:"primary_key"`
	CreateDate     string `gorm:"index"`
	FailTimes      uint32
	NoticeTime     string `gorm:"index"`
	Status         string `gorm:"index"`
	ErrorMessage   string
	NextNotifyTime string `gorm:"index"`
}

func (Notice) TableName() string{
	return "pay_notice"
}

type NoticeOk struct {
	GatewayOrderId string `gorm:"primary_key"`
	CreateDate     string `gorm:"index"`
	FailTimes      uint32
	NoticeTime     string `gorm:"index"`
}

func (NoticeOk) TableName() string{
	return "pay_notice_ok"
}