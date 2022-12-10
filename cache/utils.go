package cache

import (
	"log"
	"net/http"
)

// create a group, we can accept different name for group name
func CreateGroup(groupName string, fn Getter,cacheSize int64)*Group{
	return NewGroup(groupName,cacheSize,fn)
}

// start a cache server, user will not sense it. this will only expose to peer node
func StartCacheServer(addr string, addrs[]string, mainCache *Group){
	peers := NewNetworkController(addr)
	peers.Set(addrs...)
	mainCache.RegisterPeers(peers)
	log.Println("cache is running")
	log.Fatal(http.ListenAndServe(addr[7:],peers))
}


// start a front end interaction, this address and port will be exposed to user
func StartAPIServer(apiAddr string, cache*Group){
	http.Handle("/api",http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view,err := cache.Get(key)
			if err != nil{
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type","application/octet-stream")
			w.Write(view.ByteSlice())
		}))
		log.Println("Frontend server is running at ",apiAddr)
		log.Fatal(http.ListenAndServe(apiAddr[7:],nil))
}


