package callback

import (
	"fmt"
	"net/url"
	"testing"
)

func TestBuildChannelRequest(t *testing.T) {
	parse, err := url.Parse("?a=b&c=d")
	if err != nil{
		panic(err.Error())
	}
	fmt.Println(parse.RawQuery)
	parse, err = url.Parse("/a/b/c?a=b&c=d")
	if err != nil{
		panic(err.Error())
	}
	fmt.Println(parse.RawQuery)
	parse, err = url.Parse("/a/b/c?a=b&c=d")
	if err != nil{
		panic(err.Error())
	}
	fmt.Println(parse.Path)
	parse, err = url.Parse("/a/b/c?")
	if err != nil{
		panic(err.Error())
	}
	fmt.Println(parse.Path)
	parse, err = url.Parse("/a/b/c")
	if err != nil{
		panic(err.Error())
	}
	fmt.Println(parse.Path)
}
