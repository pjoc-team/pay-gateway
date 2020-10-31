package generator

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

var clusterID = "06"

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

func TestGenerateOrderID(t *testing.T) {
	generator := New(clusterID, 1000)
	generator.Debug()
	id := generator.GenerateID()
	fmt.Println(id)
}

var generator = New(clusterID, 1000)

func BenchmarkGenerateOrderID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generator.GenerateID()
	}
}
