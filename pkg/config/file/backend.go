package file

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/pjoc-team/pay-gateway/pkg/config/types"
	"github.com/pjoc-team/tracing/logger"
	"github.com/spf13/viper"
	"strings"
)

const (
	provider = types.Provider("file")
)

var (
	decoderConfigOptions = []viper.DecoderConfigOption{
		func(decoderConfig *mapstructure.DecoderConfig) {
			decoderConfig.ZeroFields = false // 如果没找到配置，禁止初始化
		},
	}
)

func init() {
	err := types.RegisterBackendOrDie(
		provider,
		initBackend,
		types.WithDemoURL("file://./testdata/config.yaml"),
		types.TestDemoURL(),
	)
	if err != nil {
		panic(err.Error())
	}
}

type fileBackend struct {
	config *types.Config
	v      *viper.Viper

	decoderConfigOption []viper.DecoderConfigOption
	configType          string `validate:"required"`
}

func (f *fileBackend) Start() error {
	log := logger.Log()
	err := f.v.ReadInConfig()
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	f.decoderConfigOption = append(viperDecoderConfig(f.configType), decoderConfigOptions...)
	f.v.WatchConfig()
	return nil
}

func (f *fileBackend) UnmarshalGetConfig(ctx context.Context, ptr interface{}, keys ...string) error {
	if len(keys) == 0 {
		return f.v.Unmarshal(ptr, f.decoderConfigOption...)
	}
	b := strings.Builder{}
	delimiter := ""
	for _, key := range keys {
		b.WriteString(delimiter)
		b.WriteString(key)
		delimiter = "."
	}
	k := b.String()
	return f.v.UnmarshalKey(k, ptr, f.decoderConfigOption...)
}

func initBackend(config *types.Config) (types.Backend, error) {
	f := &fileBackend{}
	v := viper.New()
	f.v = v
	filePath := fmt.Sprintf("%s%s", config.Host, config.Path)
	v.SetConfigFile(filePath)
	configType := configType(filePath)
	if configType == "" {
		return nil, fmt.Errorf("unknown file extension of file: %v", filePath)
	}
	f.config = config
	f.configType = configType
	v.SetConfigType(configType)

	vld := validator.New()
	err := vld.Struct(f)
	if err != nil {
		panic(fmt.Sprintf("validate backend error: %v", err.Error()))
	}

	return f, nil
}

func viperDecoderConfig(configType string) []viper.DecoderConfigOption {
	switch configType {
	case "yaml", "yml":
		return []viper.DecoderConfigOption{
			func(config *mapstructure.DecoderConfig) {
				config.TagName = "yaml"
			},
		}
	case "json", "toml", "properties", "props", "prop", "hcl", "dotenv", "env", "ini":
		return []viper.DecoderConfigOption{
			func(config *mapstructure.DecoderConfig) {
				config.TagName = configType
			},
		}
	}
	return nil
}

func configType(filePath string) string {
	fileIndex := strings.LastIndex(filePath, "/")
	if fileIndex >= 0 {
		filePath = filePath[fileIndex:]
	}
	index := strings.LastIndex(filePath, ".")
	if index < 0 {
		return ""
	}
	return filePath[index+1:]
}
