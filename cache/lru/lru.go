package lru

import "container/list"

type Cache struct{
	maxBytes int64 //max cache capacity in bytes
	nBytes int64 // used capacity
	ll *list.List // a double linked list
	cache map[string]*list.Element // hash map to store key and linkedlist node
	onEvicted func(key string,value Value) // onEvicted function
}

// entry is the value we store in linked list element
type entry struct{
	key string
	value Value
}

// we have a value interface
// every thing that we stored in cache must have len
type Value interface{
	Len() int
}
// constructor
func New(maxBytes int64, onEvicted func(key string, value Value))*Cache{
	return &Cache{
		maxBytes: maxBytes,
		ll: list.New(),
		cache: make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

// method the get value with key
func(c *Cache)Get(key string)(value Value,ok bool){
	// if we can find it in cache, move it to the front of the ll
	if ele,ok := c.cache[key];ok{
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value,true
	}
	return
}

// remove oldest node
func(c *Cache)RemoveOldest(){
	ele := c.ll.Back()
	if ele != nil{
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache,kv.key)
		// update size
		c.nBytes -= (int64(len(kv.key))+ int64(kv.value.Len()))
		// trigger onEvicted function
		if c.onEvicted != nil{
			c.onEvicted(kv.key,kv.value)
		}
	}
}

// add new kv into cache
func(c *Cache)Add(key string,value Value){
	// if current key existed
	if ele,ok := c.cache[key];ok{
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		// update size and value
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	}else{
		// add new entry
		ele := c.ll.PushFront(&entry{key: key,value: value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	// if size is reached, remove oldest entry
	for c.maxBytes != 0 && c.maxBytes < c.nBytes{
		c.RemoveOldest()
	}
}


