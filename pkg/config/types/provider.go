package types

import "fmt"

var (
	backends = make(map[Provider]*InitOpts)
)

// Provider backend provider
type Provider string

// InitOpts init options
type InitOpts struct {
	InitFunc InitFunc
	Options  *Options
}

// GetBackend 获取provider对应的实现
func GetBackend(provider Provider) (*InitOpts, error) {
	bf, ok := backends[provider]
	if !ok {
		return nil, fmt.Errorf("not found backend type: %v, do not forget to import the backend", provider)
	}
	return bf, nil
}
