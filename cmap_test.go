package cmap

import (
	"fmt"
	"sync"
	"testing"
)

func TestConcurrencyMap(t *testing.T) {
	cmap := NewConcurrencyMap()
	err := cmap.Set("foo", "bar")
	if err != nil {
		t.Error("Set error:", err)
		return
	}
	val, err := cmap.Get("foo")
	if err != nil {
		t.Error("Get error:", err)
		return
	}
	t.Log("Foo value:", val)
	t.Log("Map value:", cmap.ToMap(), ",Map len:", cmap.Len())
}

func BenchmarkMap(b *testing.B) {
	b.StopTimer()
	mp := make(map[interface{}]interface{})
	var lock sync.RWMutex
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		go func(i int) {
			lock.Lock()
			mp[fmt.Sprintf("foo_%d", i)] = i
			lock.Unlock()
		}(i)
	}
}

func BenchmarkConcurrencyMap(b *testing.B) {
	b.StopTimer()
	cmap := NewConcurrencyMap()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		go func(i int) {
			err := cmap.Set(fmt.Sprintf("foo_%d", i), i)
			if err != nil {
				b.Error(err)
			}
		}(i)
	}
}
