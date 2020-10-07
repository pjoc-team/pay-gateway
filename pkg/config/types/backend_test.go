package types

import (
	"fmt"
	"testing"
)

func TestParseConfig(t *testing.T) {
	config, err := ParseConfig("file://conf/test.yaml?hello=zhangsan")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("%#v \n", config)
	config, err = ParseConfig("file://./test.yaml?hello=zhangsan")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("%#v \n", config)
	config, err = ParseConfig("file://test.yaml?hello=zhangsan")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("%#v \n", config)
}
