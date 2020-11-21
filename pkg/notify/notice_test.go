package notify

import (
	"fmt"
	"github.com/pjoc-team/pay-gateway/pkg/date"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNextTimeToNotice(t *testing.T) {
	for i := 0; i < len(DEFAULT_NOTICE_EXPRESSIONG); i++ {
		nextTime, err := NextTimeToNotice(uint32(i), DEFAULT_NOTICE_EXPRESSIONG)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		fmt.Printf("now: %v nextTime: %v \n", date.NowTime(), nextTime)
		assert.Equal(t, DEFAULT_NOTICE_EXPRESSIONG[i], TimeDelayNow(nextTime))
	}
	_, err := NextTimeToNotice(uint32(len(DEFAULT_NOTICE_EXPRESSIONG)), DEFAULT_NOTICE_EXPRESSIONG)
	if err == nil {
		assert.Fail(t, "Must error!")
	} else {
		fmt.Println(err.Error())
	}

}

func TimeDelayNow(timeStr string) int {
	parse, _ := time.Parse(date.TIME_FORMAT, timeStr)
	now, _ := time.Parse(date.TIME_FORMAT, date.NowTime())
	duration := parse.Unix() - now.Unix()
	return int(duration)
}
