package notify

import (
	"encoding/json"
	"errors"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
)

// JSONMessageSerializer json serializer
type JSONMessageSerializer struct {
}

// NewJSONMessageSerializer create json serializer
func NewJSONMessageSerializer() *JSONMessageSerializer {
	return &JSONMessageSerializer{}
}

// Serialize serialize
func (*JSONMessageSerializer) Serialize(notify *pay.PayNotify) (string, error) {
	bytes, e := json.Marshal(notify)
	if e != nil {
		return "", e
	}
	return string(bytes), nil
}

// Deserialize deserialize
func (*JSONMessageSerializer) Deserialize(str string) (notify *pay.PayNotify, err error) {
	log := logger.Log()

	if str == "" {
		err = errors.New("string is empty")
		return
	}
	notify = &pay.PayNotify{}
	if err = json.Unmarshal([]byte(str), notify); err != nil {
		log.Errorf("Failed to unmarshal string: %v error: %v", str, err.Error())
		return
	}
	return
}
