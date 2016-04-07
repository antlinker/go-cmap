package cmap

import (
	"fmt"

	"testing"
)

func TestCMap(t *testing.T) {
	cmap := NewConcurrencyMap()
	cmap.Set("Foo", "bar")
	cmap.Set("Foo1", "bar1")
	cmap.Set("Foo2", "bar2")
	cmap.SetIfAbsent("Foo", "bar2")
	foo, err := cmap.Get("Foo")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Foo:", foo)
	t.Log("Keys:", cmap.Keys())
	t.Log("Values:", cmap.Values())
	t.Log("Map:", cmap.ToMap(), ",Len:", cmap.Len())
	foo2, err := cmap.Remove("Foo2")
	t.Log("Remove value:", foo2)
	t.Log("Map:", cmap.ToMap(), ",Len:", cmap.Len())
}

func TestCMapElements(t *testing.T) {
	cmap := NewConcurrencyMap()
	for i := 0; i < 10; i++ {
		err := cmap.Set(fmt.Sprintf("Foo_%d", i), i)
		if err != nil {
			t.Error(err)
			return
		}
	}
	for element := range cmap.Elements() {
		t.Log("Key:", element.Key, ",Value:", element.Value)
	}
}

func TestCMapSetGet(t *testing.T) {
	var keys []string
	keyValue := "test"
	for i := 1; i <= 10; i++ {
		keyValue = fmt.Sprintf("%s%d", keyValue, i)
		keys = append(keys, keyValue)
	}
	fmt.Println(keys)
	cmap := NewConcurrencyMap()
	for _, v := range keys {
		cmap.Set(v, v)
	}
	for _, v := range keys {
		val, _ := cmap.Get(v)
		if v != val.(string) {
			t.Error("Not the desired value:", v, val)
			return
		}
	}
}
