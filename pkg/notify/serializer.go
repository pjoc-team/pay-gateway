package notify

import (
	"encoding/json"
	"errors"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
)

type JsonMessageSerializer struct {
}

func NewJsonMessageSerializer() *JsonMessageSerializer{
	return &JsonMessageSerializer{}
}

func (*JsonMessageSerializer) Serialize(notice pay.PayNotice) (string, error) {
	if bytes, e := json.Marshal(notice); e != nil {
		return "", e
	} else {
		return string(bytes), e
	}
}

func (*JsonMessageSerializer) Deserialize(str string) (notice *pay.PayNotice, err error) {
	log := logger.Log()

	if str == "" {
		err = errors.New("string is empty")
		return
	}
	notice = &pay.PayNotice{}
	if err = json.Unmarshal([]byte(str), notice); err != nil {
		log.Errorf("Failed to unmarshal string: %v error: %v", str, err.Error())
		return
	}
	return
}
