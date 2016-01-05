package cmap

import (
	"bytes"
	"encoding/binary"
	"hash/fnv"
	"sync"
)

const (
	// DefaultPoolSize 提供分配共享池大小的默认值
	DefaultPoolSize = 1 << 5
)

// ConcurrencyMap 并发的Map接口
type ConcurrencyMap interface {
	// Get 获取给定键值对应的元素值。若没有对应的元素值则返回nil
	Get(key interface{}) (interface{}, error)
	// Set 给指定的键设置元素值。若该键值已存在，则替换
	Set(key interface{}, elem interface{}) error
	// SetIfAbsent 给指定的键设置元素值。若该键值已存在，则不替换,并返回以存在的值
	//返回值 value 为执行方法后key 键对应的实际值
	//返回值isnew true是新值，false是原来的值

	SetIfAbsent(key interface{}, elem interface{}) (value interface{}, isnew bool)
	// Remove 删除给定键值对应的键值对，并返回旧的元素值。若没有旧元素的值则返回nil
	Remove(key interface{}) (interface{}, error)
	// Contains 判断是否包含给定的键值
	Contains(key interface{}) (bool, error)
	// ToMap 获取已包含的键值对所组成的字典值
	ToMap() map[interface{}]interface{}
	// Clear 清除所有的键值对
	Clear()
	// Len 获取键值对的数量
	Len() int
}

// NewConcurrencyMap 创建并发的Map接口
// poolSize 分配共享池的大小，默认为32
func NewConcurrencyMap(poolSizes ...uint) ConcurrencyMap {
	var size uint
	if len(poolSizes) > 0 {
		size = poolSizes[0]
	} else {
		size = DefaultPoolSize
	}
	pools := make([]*concurrencyItem, size)
	for i := 0; i < int(size); i++ {
		pools[i] = &concurrencyItem{
			items: make(map[interface{}]interface{}),
		}
	}
	return &concurrencyMap{
		size:  int(size),
		pools: pools,
	}
}

type concurrencyItem struct {
	sync.RWMutex
	items map[interface{}]interface{}
}

type concurrencyMap struct {
	size  int
	pools []*concurrencyItem
}

func (cm *concurrencyMap) getItem(key interface{}) (*concurrencyItem, error) {
	var p []byte
	switch key.(type) {
	case []byte:
		p = key.([]byte)
	case string:
		p = []byte(key.(string))
	default:
		buffer := new(bytes.Buffer)
		err := binary.Write(buffer, binary.LittleEndian, key)
		if err != nil {
			return nil, err
		}
		p = buffer.Bytes()
	}
	hasher := fnv.New32()
	_, err := hasher.Write(p)
	if err != nil {
		return nil, err
	}
	return cm.pools[uint(hasher.Sum32())%uint(cm.size)], nil
}

func (cm *concurrencyMap) Get(key interface{}) (interface{}, error) {
	item, err := cm.getItem(key)
	if err != nil {
		return nil, err
	}
	item.RLock()
	v := item.items[key]
	item.RUnlock()
	return v, nil
}

func (cm *concurrencyMap) Set(key interface{}, elem interface{}) error {
	item, err := cm.getItem(key)
	if err != nil {
		return err
	}
	item.Lock()
	item.items[key] = elem
	item.Unlock()
	return nil
}

func (cm *concurrencyMap) SetIfAbsent(key interface{}, elem interface{}) (interface{}, bool) {
	item, err := cm.getItem(key)
	if err != nil {
		return item, false
	}
	item.Lock()
	_, ok := item.items[key]
	if !ok {
		item.items[key] = elem
	}
	item.Unlock()
	return elem, true
}

func (cm *concurrencyMap) Remove(key interface{}) (interface{}, error) {
	item, err := cm.getItem(key)
	if err != nil {
		return nil, err
	}
	item.Lock()

	elem, ok := item.items[key]
	if ok {
		delete(item.items, key)
	}
	item.Unlock()
	return elem, nil
}

func (cm *concurrencyMap) Contains(key interface{}) (bool, error) {
	item, err := cm.getItem(key)
	if err != nil {
		return false, err
	}
	item.RLock()

	_, ok := item.items[key]
	item.RUnlock()
	return ok, nil
}

func (cm *concurrencyMap) ToMap() map[interface{}]interface{} {
	data := make(map[interface{}]interface{})
	for i := 0; i < cm.size; i++ {
		item := cm.pools[i]
		item.RLock()
		for k, v := range item.items {
			data[k] = v
		}
		item.RUnlock()
	}
	return data
}

func (cm *concurrencyMap) Clear() {
	for i := 0; i < cm.size; i++ {
		item := cm.pools[i]
		item.Lock()
		item.items = make(map[interface{}]interface{})
		item.Unlock()
	}
}

func (cm *concurrencyMap) Len() int {
	var count int
	for i := 0; i < int(cm.size); i++ {
		item := cm.pools[i]
		item.RLock()
		count += len(item.items)
		item.RUnlock()
	}
	return count
}
