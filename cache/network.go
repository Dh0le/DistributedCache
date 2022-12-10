package cache

import (
	"cache/consistenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// this path will be used in node communication
const defaultBasePath = "/_gocache/"
const defaultReplicas = 5
var _PeerPicker = (*NetworkController)(nil)
var _PeerGetter = (*httpGetter)(nil)


// HttpPool is a struct implementted hanlder, PeerPicker interface
type NetworkController struct{
	self string // address and port for current node
	basePath string // base url for cache api
	mu sync.Mutex // mutex lock for register peer
	peers *consistenthash.Map // a consistant hash object to add and map peers
	httpGetters map[string]*httpGetter // a hash map that map peer name to its getter function
}

// consturctor of HTTPPool
func NewNetworkController(self string)*NetworkController{
	return &NetworkController{
		self: self,
		basePath: defaultBasePath,
	}
}

// Log function
func(p *NetworkController)Log(format string ,v ...interface{}){
	log.Printf("[Server %s]%s",p.self,fmt.Sprintf(format,v...))
}

// ServeHTTP function to implement Handler interface
func(p *NetworkController)ServeHTTP(w http.ResponseWriter,r *http.Request){
	// check if current request has correct path
	if !strings.HasPrefix(r.URL.Path,p.basePath){
		panic("HTTPPOOL serving unexpected path:" + r.URL.Path)
	}
	p.Log("%s,%s",r.Method,r.URL.Path)
	// split path to get group and key name
	parts := strings.SplitN(r.URL.Path[len(p.basePath):],"/",2)
	if len(parts) != 2{
		http.Error(w,"bad request",http.StatusBadRequest)
		return 
	}
	groupName := parts[0]
	key := parts[1]

	// get group in cache
	group := GetGroup(groupName)
	if group == nil{
		http.Error(w,"no such group",http.StatusNotFound)
		return
	}
	
	// fetch data in current group
	view, err := group.Get(key)
	if err != nil{
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	// write response
	w.Header().Set("Content-Type","application/octet-stream")
	w.Write(view.ByteSlice())
}

// function to set peers for current node
func (p *NetworkController)Set(peers ...string){
	// lock to prevent conflict
	p.mu.Lock()
	defer p.mu.Unlock()
	// create consistant hash object
	p.peers = consistenthash.New(defaultReplicas,nil)
	// add all peers into the hashring and create mapping
	p.peers.Add(peers...)
	// for each peer, we create its mapping between its name and its getter function
	p.httpGetters = make(map[string]*httpGetter,len(peers))
	// create getter function for each peer
	// the base url for the getter function is the name of the peer with base path
	for _,peer:= range peers{
		p.httpGetters[peer] = &httpGetter{baseUrl: peer+p.basePath}
	}
}

// function to implement PeerPicker interface, then we can inject this object into our maincache
func(p *NetworkController)PickPeer(key string)(PeerGetter,bool){
	// lock to prevent conflict
	p.mu.Lock()
	defer p.mu.Unlock()
	// in the consistant hash , we find the peer that store the val of given key
	if peer := p.peers.Get(key);peer != "" && peer != p.self{
		// if peer is found, return its getter function
		p.Log("Pick peer %s",peer)
		return p.httpGetters[peer],true
	}
	return nil,false
}

// a getter object to retrieve data from peer node(Implemented peerGetter interface)
type httpGetter struct{
	baseUrl string
}

func(h *httpGetter)Get(group string, key string)([]byte,error){
	// create the url for peer node
	u := fmt.Sprintf(
		"%v%v/%v",h.baseUrl,url.QueryEscape(group),url.QueryEscape(key),
	)
	// send get request
	 res, err := http.Get(u)
	// fetch failed
	 if err != nil{
		return nil,err
	 }
	 // check status
	 if res.StatusCode != http.StatusOK{
		return nil,fmt.Errorf("server returned %v",res.Status)
	 }
	 // parse body
	 bytes, err := ioutil.ReadAll(res.Body)
	 // parse error
	 if err != nil{
		return nil,fmt.Errorf("reading response body:%v",err)
	 }
	 // successfully fetched
	 return bytes,nil

}