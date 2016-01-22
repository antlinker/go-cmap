package cmap_test

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	. "github.com/antlinker/go-cmap"
)

func TestGetElement(t *testing.T) {
	cmap := NewConcurrencyMap()
	cmap.Set("foo", "bar")
	for ele := range cmap.GetElement() {
		t.Log("Element:", ele.Key, ele.Value)
	}
	t.Log("Success")
}

type TestMap struct {
	gomap       map[interface{}]interface{}
	cmap        ConcurrencyMap
	gomapnolock map[interface{}]interface{}
	sync.RWMutex
}

func NewTestMap() *TestMap {
	return &TestMap{gomap: make(map[interface{}]interface{}), cmap: NewConcurrencyMap(), gomapnolock: make(map[interface{}]interface{}, 1024*1024)}
}
func (m *TestMap) GomapNolockGetSet(key interface{}, value interface{}) bool {

	m.gomap[key] = value

	runtime.Gosched()

	newvalue := m.gomap[key]

	return newvalue == value

}
func (m *TestMap) GomapGetSet(key interface{}, value interface{}) bool {
	m.Lock()
	m.gomap[key] = value
	m.Unlock()
	runtime.Gosched()
	m.RLock()
	newvalue := m.gomap[key]
	m.RUnlock()
	return newvalue == value

}
func (m *TestMap) ConcurrencymapGetSet(key interface{}, value interface{}) bool {
	err := m.cmap.Set(key, value)
	if err != nil {
		return false
	}
	newvalue, _ := m.cmap.Get(key)
	return newvalue == value
}

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

func BenchmarkNolockGoMap(b *testing.B) {
	b.StopTimer()
	testmap := NewTestMap()
	b.StartTimer()
	var i int64 = 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := atomic.AddInt64(&i, 1)
			var key = fmt.Sprintf("foo_%d", n)
			result := testmap.GomapGetSet(key, n)
			if !result {
				b.Error("执行错误错误结果")
			}
		}
	})
}

func BenchmarkGoMap(b *testing.B) {
	b.StopTimer()
	testmap := NewTestMap()
	b.StartTimer()
	var i int64 = 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := atomic.AddInt64(&i, 1)
			var key = fmt.Sprintf("foo_%d", n)
			result := testmap.GomapGetSet(key, n)
			if !result {
				b.Error("执行错误错误结果")
			}
		}
	})
}

func BenchmarkConcurrencyMap(b *testing.B) {
	b.StopTimer()
	testmap := NewTestMap()
	b.StartTimer()
	var i int64 = 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := atomic.AddInt64(&i, 1)
			var key = fmt.Sprintf("foo_%d", n)
			result := testmap.ConcurrencymapGetSet(key, n)
			if !result {
				b.Error("执行错误错误结果")
			}
		}
	})

}
