package main

import (
	"cache"
	"flag"
	"fmt"
	"log"
	"strconv"
)

// dummy db
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main(){
	// allowed user to decide if we want to start a api server
	var port int
	var api bool
	flag.IntVar(&port,"port",8001,"Cache server port")
	flag.BoolVar(&api,"api",false,"Start a api server?")
	flag.Parse()

	// so there is where you place your 
	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001:"http://localhost:8001",
		8002:"http://localhost:8002",
		8003:"http://localhost:8003",
	}

	var addrs []string

	for _,v := range addrMap{
		addrs = append(addrs, v)
	}

	// here is only a dummy getter function
	// place logic with database here
	getterFn := cache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key:",key)
		if v,ok := db[key];ok{
			return []byte(v),nil;
		}
		return nil,fmt.Errorf("%s not exist",key)
	})
	// create a group
	cacheGroup := cache.CreateGroup("scores",getterFn,2<<10)
	// create an API server
	if api{
		// since we use gin as our sever, we need to use "go" to start a new thread
		// if we dont use go here, the thread will stuck and will not proceed to create local cache server
		go cache.StartAPIServer(apiAddr,":9999",cacheGroup)
	}
	
	// start Cache server
	cache.StartCacheServer(addrMap[port],":"+ strconv.Itoa(port),addrs,cacheGroup)
}

