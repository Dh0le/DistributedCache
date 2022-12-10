package main

import (
	"cache"
	"flag"
	"fmt"
	"log"
)

// dummy db
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// // start a cache server, user will not sense it. this will only expose to peer node
// func startCacheServer(addr string, addrs[]string, mainCache *cache.Group){
// 	peers := cache.NewHTTPPool(addr)
// 	peers.Set(addrs...)
// 	mainCache.RegisterPeers(peers)
// 	// will be modified use Gin framework
// 	log.Println("cache is running")
// 	log.Fatal(http.ListenAndServe(addr[7:],peers))
// }

// // start a front end interaction, this address and port will be exposed to user
// // TODO: This wll be modifed to use Gin framework
// func startAPIServer(apiAddr string, cache*cache.Group){
// 	http.Handle("/api",http.HandlerFunc(
// 		func(w http.ResponseWriter, r *http.Request) {
// 			key := r.URL.Query().Get("key")
// 			view,err := cache.Get(key)
// 			if err != nil{
// 				http.Error(w,err.Error(),http.StatusInternalServerError)
// 				return
// 			}
// 			w.Header().Set("Content-Type","application/octet-stream")
// 			w.Write(view.ByteSlice())
// 		}))
// 		log.Println("Frontend server is running at ",apiAddr)
// 		log.Fatal(http.ListenAndServe(apiAddr[7:],nil))
// 	}


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

	cacheGroup := cache.CreateGroup("score",getterFn,2<<10)
	if api{
		cache.StartAPIServer(apiAddr,cacheGroup)
	}

	cache.StartCacheServer(addrMap[port],addrs,cacheGroup)
}

