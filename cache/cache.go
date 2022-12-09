package cache

import (
	"cache/lru"
	"sync"
)

// the cache it self is concurrent
// inorder to secure the data, we wrap the lru cache with a thread safe struct cache
type cache struct{
	// lock for thread safety
	mu sync.Mutex
	lru *lru.Cache
	cacheByte int64
}

// add new kv into lru cache
func(c *cache)add(key string, value ByteView){
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil{
		c.lru = lru.New(c.cacheByte,nil)
	}
	c.lru.Add(key,value)
}
// get value from lru
func(c *cache)get(key string)(value ByteView,ok bool){
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil{
		return
	}
	if v,ok := c.lru.Get(key);ok{
		return v.(ByteView),ok
	}
	return 
}

