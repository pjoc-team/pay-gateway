package main

import "fmt"

type A struct {
	kv map[string]string
}

type B struct {
	a *A
}

func main() {
	a := &A{}
	fmt.Println(a.kv["aaa"])
	if a.kv == nil {
		a.kv = make(map[string]string)
	}
	a.kv["aaa"] = "v"
	fmt.Println(a.kv["aaa"])
}
