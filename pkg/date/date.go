package date

import "time"

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
	DATE_FORMAT = "2006-01-02"
)

func NowDate() string {
	return time.Now().Format(DATE_FORMAT)
}

func NowTime() string {
	return time.Now().Format(TIME_FORMAT)
}
