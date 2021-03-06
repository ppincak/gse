package socket

import (
	"sync"
)

type LocalStore struct {
	data	map[string] interface{}
	mtx     *sync.RWMutex
}

func NewLocalStore() Store {
	return &LocalStore{
		data: 	make(map[string] interface{}),
		mtx: 	new(sync.RWMutex),
	}
}

func (client *LocalStore) Set(key string, value interface{}) {
	client.mtx.Lock()
	client.data[key] = value
	client.mtx.Unlock()
}

func (client *LocalStore) Get(key string) interface{} {
	client.mtx.RLock()
	val := client.data[key]
	client.mtx.RUnlock()
	return val
}

func (client *LocalStore) Delete(key string) {
	client.mtx.Lock()
	delete(client.data, key)
	client.mtx.Unlock()
}

func (client *LocalStore) DeleteAndGet(key string) interface {} {
	client.mtx.Lock()
	defer client.mtx.Unlock()
	val := client.data[key]
	delete(client.data, key)
	return val
}

func (client *LocalStore) Has(key string) bool {
	client.mtx.RLock()
	_, ok := client.data[key]
	client.mtx.RUnlock()
	return ok
}

func (client *LocalStore) Destroy() {
	client.data = make(map[string]interface{});
}