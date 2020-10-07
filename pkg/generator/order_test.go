package generator

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

var clusterId = "06"

func TestDate(t *testing.T) {
	fmt.Println(math.MaxInt64)
	n := 2020*100 + 7 // month
	n = n*100 + 20    // day
	n = n*100 + 9     // hour
	n = n*100 + 2     // minutes
	n = n*100 + 0     // seconds
	n = n*1000 + 3    // seconds
	fmt.Println(n)
	fmt.Println(reflect.TypeOf(n))
}

func TestGenerateOrderId(t *testing.T) {
	generator := New(clusterId, 1000)
	generator.Debug()
	id := generator.GenerateId()
	fmt.Println(id)
}

var generator = New(clusterId, 1000)

func BenchmarkGenerateOrderId(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generator.GenerateId()
	}
}
