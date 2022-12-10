package cache

import (
	"fmt"
	"log"
	"sync"
)

// Getter interface, since for different group we will have different data getter function
// These function implement Getter function allows different databases support
type Getter interface{
	Get(key string)([]byte,error)
}

// we create a getter function to implement Getter interface
type GetterFunc func(key string)([]byte,error)

func (f GetterFunc)Get(key string)([]byte,error){
	return f(key)
}

var(
	// we have readwrite lock when user creating new groups
	mu sync.RWMutex
	// we have hash map for all the groups with its name as key
	groups = make(map[string]*Group)
)

// group is the top granularity of the cache. Same type of data will be stored in the same group like"Score","Rating"
type Group struct{
	name string // name of the group
	getter Getter // getter that allows cache to fetch data from database, this is a cache through architecture
	mainCache cache // concurrent hash for this group in this node(this is a distributed cache)
	peers PeerPicker // add pick peer that allows data to be fetched from other nodes
}

// Constructor for a group
func NewGroup(name string, cacheBytes int64, getter Getter)*Group{
	// every gourp must have a getter
	if getter == nil{
		panic("nil getter")
	}
	// when createing a group, we have a mutex lock so there is no conflict
	mu.Lock()
	defer mu.Unlock()
	// create group
	g := &Group{
		name: name,
		getter: getter,
		mainCache: cache{cacheByte: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string)*Group{
	// this is ready only ,so we add read lock here
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

// get value with given key in this group
func (g *Group)Get(key string)(ByteView,error){
	// key must not be empty or null
	if key == ""{
		return ByteView{},fmt.Errorf("key is required")
	}
	// if cache in current node contains the value
	if v,ok := g.mainCache.get(key);ok{
		log.Println("Cache hit")
		return v,nil
	}
	// current node does not corresponding value
	// entering remote fetching process
	return g.load(key)
}

// for now we just fetch from database directly.
// later we will add fetch process to fetch data from peer node
func (g *Group)load(key string)(ByteView,error){
	return g.getLocally(key)
}

func (g *Group)getFromPeer(peer PeerGetter,key string)(ByteView,error){
	bytes,err := peer.Get(g.name,key)
	if err != nil{
		return ByteView{},err
	}
	return ByteView{bytes},nil
}

// fetch data from database
func (g *Group)getLocally(key string)(ByteView,error){
	bytes,err := g.getter.Get(key)
	// fetch failed
	if err != nil{
		return ByteView{},err
	}
	// retrieve the data
	value := ByteView{
		// return its copy, its read only
		b:cloneByte(bytes),
	}
	// place fetched value and key into cache in this node
	// Explaination:
	// At the first time I was thinking if this will violate the rules of consistant hashing.
	// then I realized that in this case, we are actually using a distributed database as well
	// So if the key are not suppose to be store in this cache. Other cache node should have trigger this procedure as well.
	// Only if other peer node does not have the data, then we will reach this step.
	g.populateCache(key,value)
	return value,nil
}

// add node and value into cache in current node
func (g *Group)populateCache(key string,value ByteView){
	g.mainCache.add(key,value)
}






