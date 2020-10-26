package config

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pjoc-team/pay-gateway/pkg/config/types"
	"github.com/pjoc-team/pay-gateway/pkg/config/types/mock"
	"gopkg.in/go-playground/assert.v1"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"testing"
)

func TestDefaultServer_GetConfig(t *testing.T) {
	type person struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}

	controller := gomock.NewController(t)
	backend := mock.NewMockBackend(controller)
	var data = map[string]map[string]interface{}{}
	err := types.RegisterBackendOrDie(
		"mock", func(config *types.Config) (types.Backend, error) {
			filePath := fmt.Sprintf("%s%s", config.Host, config.Path)
			file, err := ioutil.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			err = yaml.Unmarshal(file, data)
			fmt.Println("data:", data)
			if err != nil {
				return nil, err
			}
			return backend, nil
		},
	)
	p := &person{}
	backend.EXPECT().UnmarshalGetConfig(p, "appID1", "mchID1").Do(
		func(p interface{}, keys ...string) {
			if len(keys) != 2 {
				return
			}
			data := data[keys[0]][keys[1]]
			marshal, err := yaml.Marshal(data)
			if err != nil {
				t.Fatal(err.Error())
			}
			err = yaml.Unmarshal(marshal, p)
			if err != nil {
				t.Fatal(err.Error())
			}
		},
	).Return(nil).MinTimes(1)
	if err != nil {
		t.Fatal(err.Error())
	}

	server, err := InitConfigServer("mock://./test.yaml?hello=zhangsan")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("%#v\n", data)

	err = server.UnmarshalGetConfig(context.Background(), p, "appID1", "mchID1")
	if err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, p.Name, "zhangsan")
	assert.Equal(t, p.Age, 18)
}
