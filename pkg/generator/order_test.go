package generator

import (
	"fmt"
	"gopkg.in/go-playground/assert.v1"
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

func TestGenerator_GenerateIndex(t *testing.T) {
	type fields struct {
		ClusterID           string
		MachineID           string
		ClusterAndMachineID string
		Concurrency         int
		maxIndex            int32
		indexWidth          int
		index               int32
		byteLength          int
		debug               bool
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name:   "t1",
			fields: fields{
				ClusterID:           "t1",
				MachineID:           "m1",
				ClusterAndMachineID: "",
				Concurrency:         1,
				maxIndex:            99,
				indexWidth:          2,
				index:               0,
				byteLength:          0,
				debug:               false,
			},
			want:   []byte("01"),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				g := &Generator{
					ClusterID:           tt.fields.ClusterID,
					MachineID:           tt.fields.MachineID,
					ClusterAndMachineID: tt.fields.ClusterAndMachineID,
					Concurrency:         tt.fields.Concurrency,
					maxIndex:            tt.fields.maxIndex,
					indexWidth:          tt.fields.indexWidth,
					index:               tt.fields.index,
					byteLength:          tt.fields.byteLength,
					debug:               tt.fields.debug,
				}
				if got := g.GenerateIndex(); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GenerateIndex() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestGenerator_GenerateIndex1(t *testing.T) {
	g := New("t1", 1000000)
	index := g.GenerateIndex()
	assert.Equal(t, index, []byte("0000001"))
}