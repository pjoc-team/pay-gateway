package discovery

import (
	"context"
	"fmt"
	"github.com/prometheus/common/log"
	"sync"
	"testing"
)

func TestFileStore_Put(t *testing.T) {
	store, err := NewFileStore(context.Background(), "./test.data")
	if err != nil {
		t.Fatal(err.Error())
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(ii int) {
			defer wg.Done()
			s := &Service{
				ServiceName: fmt.Sprintf("test%d", ii),
				IP:          fmt.Sprintf("192.168.0.%d", ii),
				Port:        ii,
			}
			err = store.Put(s.ServiceName, s)
			if err != nil {
				log.Error(err.Error())
			}
		}(i)
	}
	wg.Wait()
}
