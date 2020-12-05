package main

import "fmt"


func main() {
	// A demo struct
	type A struct {
		kv map[string]string
	}
	a := &A{}
	fmt.Println(a.kv["aaa"])
	if a.kv == nil {
		a.kv = make(map[string]string)
	}
	a.kv["aaa"] = "v"
	fmt.Println(a.kv["aaa"])
}
