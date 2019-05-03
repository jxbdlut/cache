package cache

import (
	"sort"
	"sync"
)

var cacheConnsMap  = make(map[string]CacheRedis)
var index int
var connMutex sync.Mutex

// 设置 cache 地址
func (c *Cache) AddCacheAddress(address, password string) {
	connMutex.Lock()
	defer connMutex.Unlock()
	conn := NewRedisCache(address, password, 0)
	cacheConnsMap[address] = conn
}

func (c *Cache) SetCacheAddress(host map[string]string) {
	connMutex.Lock()
	defer connMutex.Unlock()
	cacheConnsMap = make(map[string]CacheRedis)
	for address, password := range host {
		cacheConnsMap[address] = NewRedisCache(address, password, 0)
	}
}

func (c *Cache) DelCacheAddress(addrs []string) {
	connMutex.Lock()
	defer connMutex.Unlock()
	for _, addr := range addrs {
		if conn, ok := cacheConnsMap[addr]; ok {
			conn.Close()
			delete(cacheConnsMap, addr)
		}
	}
}

func (c *Cache) GetCacheClient() CacheRedis {
	connMutex.Lock()
	defer connMutex.Unlock()
	sortedKeys := make([]string, 0)
	for key := range cacheConnsMap {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	index = (index + 1) % len(sortedKeys)
	return cacheConnsMap[sortedKeys[index]]
}
