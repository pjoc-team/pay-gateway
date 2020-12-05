package notify

import (
	"fmt"
	"github.com/pjoc-team/pay-gateway/pkg/date"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNextTimeToNotice(t *testing.T) {
	for i := 0; i < len(DefaultNotifyExpression); i++ {
		nextTime, err := NextTimeToNotice(uint32(i), DefaultNotifyExpression)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		fmt.Printf("now: %v nextTime: %v \n", date.NowTime(), nextTime)
		assert.Equal(t, DefaultNotifyExpression[i], TimeDelayNow(nextTime))
	}
	_, err := NextTimeToNotice(uint32(len(DefaultNotifyExpression)), DefaultNotifyExpression)
	if err == nil {
		assert.Fail(t, "Must error!")
	} else {
		fmt.Println(err.Error())
	}

}

func TimeDelayNow(timeStr string) int {
	parse, _ := time.Parse(date.TimeFormat, timeStr)
	now, _ := time.Parse(date.TimeFormat, date.NowTime())
	duration := parse.Unix() - now.Unix()
	return int(duration)
}
