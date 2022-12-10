package cache

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// create a group, we can accept different name for group name
func CreateGroup(groupName string, fn Getter,cacheSize int64)*Group{
	return NewGroup(groupName,cacheSize,fn)
}

// start a cache server, user will not sense it. this will only expose to peer node
func StartCacheServer(addr string, port string, addrs[]string, mainCache *Group){
	r := gin.Default()
	networkController := NewNetworkController(addr)
	networkController.Set(addrs...)
	queryPath := networkController.basePath+":group/:key"
	mainCache.RegisterPeers(networkController)
	log.Println(queryPath)
	r.GET(queryPath,func(ctx *gin.Context) {
		log.Println("Recieved a fetch request from peer node")
		groupName := ctx.Param("group")
		key := ctx.Param("key")
		group := GetGroup(groupName)
		if group == nil{
			ctx.String(http.StatusNotFound,"no such group")
			return
		}
		view, err := group.Get(key)
		if err != nil{
			ctx.String(http.StatusInternalServerError,err.Error())
			return 
		}
		ctx.Header("Content-Type","application/octet-stream")
		ctx.String(http.StatusOK,view.String())
	})
	r.Run(port)
}


// start a front end interaction, this address and port will be exposed to user
func StartAPIServer(apiAddr string,port string, cache*Group){
	r := gin.Default()
	r.GET("/api",func(ctx *gin.Context) {
		key := ctx.DefaultQuery("key","Tom")
		view,err := cache.Get(key)
		if err != nil{
			ctx.String(http.StatusInternalServerError,"")
			return
		}
		ctx.Header("Content-Type","application/octet-stream")
		ctx.String(http.StatusOK,view.String())
	})
	r.Run(port)
}


