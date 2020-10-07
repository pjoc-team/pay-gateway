package file

import (
	"context"
	"fmt"
	"github.com/pjoc-team/pay-gateway/pkg/config/types"
	"gopkg.in/go-playground/assert.v1"
	"testing"
)

func Test_configType(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{
				filePath: "./foo/bar.yaml",
			},
			want: "yaml",
		},
		{
			name: "t2",
			args: args{
				filePath: "foo/bar.yaml",
			},
			want: "yaml",
		},
		{
			name: "t3",
			args: args{
				filePath: "foo/bar.json",
			},
			want: "json",
		},
		{
			name: "t3",
			args: args{
				filePath: "./bar.json",
			},
			want: "json",
		},
		{
			name: "t3",
			args: args{
				filePath: "bar.json",
			},
			want: "json",
		},
		{
			name: "t3",
			args: args{
				filePath: "bar.",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := configType(tt.args.filePath); got != tt.want {
				t.Errorf("configType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fileBackend_GetConfig(t *testing.T) {
	c := &types.Config{
		Host: ".",
		Path: "/testdata/config.yaml",
	}
	backend, err := initBackend(c)
	if err != nil {
		t.Fatal(err.Error())
	}
	type p struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}
	backend.Start()

	np := &p{}
	err = backend.UnmarshalGetConfig(context.Background(), np, "appID1", "mchID1")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(np)
	assert.Equal(t, np.Name, "zhangsan")
	assert.Equal(t, np.Age, 18)

	np2 := &p{}
	err = backend.UnmarshalGetConfig(context.Background(), np2, "appID1")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(np2)
	assert.Equal(t, np2.Name, "")
	assert.Equal(t, np2.Age, 0)
}
