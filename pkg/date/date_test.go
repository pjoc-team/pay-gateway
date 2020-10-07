package date

import (
	"fmt"
	"github.com/coreos/etcd/pkg/testutil"
	"testing"
	"time"
)

func TestNowDate(t *testing.T) {
	fmt.Println(NowDate())
	testutil.AssertEqual(t, 10, len(NowDate()))
}

func TestNowTime(t *testing.T) {
	fmt.Println(NowTime())
	testutil.AssertEqual(t, 19, len(NowTime()))

}

func TestTomorrow(t *testing.T) {
	now := time.Now()
	yesDateStr := "'" + now.AddDate(0, 0, -1).Format("2006-01-02") + "'"
	tomDateStr := "'" + now.AddDate(0, 0, 1).Format("2006-01-02") + "'"
	fmt.Println(yesDateStr)
	fmt.Println(tomDateStr)
}
