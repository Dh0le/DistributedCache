package cache

import (
	"cache/singleflight"
	"fmt"
	"log"
	"sync"
)

// Getter interface
type Getter interface{
	Get(key string)([]byte,error)
}

// Getter function interface to support dependency injection of different database
type GetterFunc func(key string)([]byte, error)

// for every getter function, return its get result
func(f GetterFunc) Get(key string)([]byte,error){
	return f(key)
}

// group is the top granularity of the cache. Same type of data will be stored in the same group like"Score","Rating"
type Group struct{
	name string // name of group
	getter Getter // getter function for current group
	mainCache cache // concurrent cache for current group
	peers PeerPicker // peer picker to fetch from peer if searched key is not in current cache
	loader *singleflight.Group // a single flight gourp to prevent cache penetration
}

var(
	mu sync.RWMutex // global RW lock for group creation
	groups = make(map[string]*Group) // global map to store all the groups
)

// constructor of a group, with name, cache size, and getter function to fetch data from database
func NewGroup(name string,cacheBytes int64, getter Getter)*Group{
	if(getter == nil){
		panic("nil getter")
	}
	// lock to prevent conflict
	mu.Lock()
	defer mu.Unlock()
	// create group
	g := &Group{
		name:name,
		getter: getter,
		mainCache: cache{cacheByte: cacheBytes},
		loader: &singleflight.Group{},
	}
	groups[name] = g
	// return created group
	return g;
}

// global function to retrieve group.
func GetGroup(name string)*Group{
	// add read lock to prevent conflict
	mu.RLock();
	defer mu.RUnlock()
	// return group
	g := groups[name]
	return g
}
// get function to get key from current group
func (g *Group) Get(key string)(ByteView,error){
	// null check for key
	if(key == ""){
		return ByteView{},fmt.Errorf("key is required")
	}
	// try to get value from cache in this node
	if v,ok := g.mainCache.get(key);ok{
		log.Println("Cache hit")
		return v,nil
	}
	// current node does not contain corresponding value
	// entering remote fetching process
	return g.load(key)
}

// function to load data from remote node or database
func (g *Group) load(key string) (value ByteView, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// use the peer getter function to fetch data
func (g *Group)getFromPeer(peer PeerGetter, key string)(ByteView,error){
	bytes,err := peer.Get(g.name,key)
	if err != nil{
		// fetch failed
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

// inject peer picker into current node
func (g *Group)RegisterPeers(peers PeerPicker){
	if g.peers != nil{
		panic("Register peers called more than once")
	}
	g.peers = peers
}



