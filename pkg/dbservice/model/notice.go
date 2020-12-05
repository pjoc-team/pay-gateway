package model

// Notify notify model
type Notify struct {
	GatewayOrderID string `gorm:"primary_key"`
	CreateDate     string `gorm:"index"`
	FailTimes      uint32
	NotifyTime     string `gorm:"index"`
	Status         string `gorm:"index"`
	ErrorMessage   string
	NextNotifyTime string `gorm:"index"`
}

// TableName table name of notify
func (Notify) TableName() string{
	return "pay_notify"
}

// NotifyOk notify ok model
type NotifyOk struct {
	GatewayOrderID string `gorm:"primary_key"`
	CreateDate     string `gorm:"index"`
	FailTimes      uint32
	NotifyTime     string `gorm:"index"`
}

// TableName table name of NotifyOk model
func (NotifyOk) TableName() string{
	return "pay_notify_ok"
}