package date

import "time"

const (
	// TimeFormat 时间格式
	TimeFormat = "2006-01-02T15:04:05.999Z07:00"
	// DateFormat 日期格式
	DateFormat = "2006-01-02"
)

// NowDate Now date string
func NowDate() string {
	return time.Now().Format(DateFormat)
}

// NowTime Now time string
func NowTime() string {
	return time.Now().Format(TimeFormat)
}
