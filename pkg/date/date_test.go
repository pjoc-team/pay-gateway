package date

import (
	"fmt"
	"gopkg.in/go-playground/assert.v1"
	"testing"
	"time"
)

func TestNowDate(t *testing.T) {
	fmt.Println(NowDate())
	assert.Equal(t, 10, len(NowDate()))
}

func TestNowTime(t *testing.T) {
	fmt.Println(NowTime())
	assert.Equal(t, 29, len(NowTime()))
}

func TestTomorrow(t *testing.T) {
	now := time.Now()
	yesDateStr := "'" + now.AddDate(0, 0, -1).Format("2006-01-02") + "'"
	tomDateStr := "'" + now.AddDate(0, 0, 1).Format("2006-01-02") + "'"
	fmt.Println(yesDateStr)
	fmt.Println(tomDateStr)
}
