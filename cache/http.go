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

const defaultBasePath = "/_geecache/"
const defaultReplicas = 50
var _PeerPicker = (*HTTPPool)(nil)
var _PeerGetter = (*httpGetter)(nil)

type HTTPPool struct{
	self string  //address and port for current node
	basePath string // baseurl path for cache api
	mu sync.Mutex // lock when registering peer
	peers *consistenthash.Map // map object to get peer with consistent hashing
	httpGetters map[string]*httpGetter
}

// consturctor of HTTPPool
func NewHTTPPool(self string)*HTTPPool{
	return &HTTPPool{
		self: self,
		basePath: defaultBasePath,
	}
}

// Log function
func(p *HTTPPool)Log(format string ,v ...interface{}){
	log.Printf("[Server %s]%s",p.self,fmt.Sprintf(format,v...))
}

// ServeHTTP function to implement Handler interface
func(p *HTTPPool)ServeHTTP(w http.ResponseWriter,r *http.Request){
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
func (p *HTTPPool)Set(peers ...string){
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
	for _,peer:= range peers{
		p.httpGetters[peer] = &httpGetter{baseUrl: peer+p.basePath}
	}
}

// function to implement PeerPicker interface, then we can inject this object into our maincache
func(p *HTTPPool)PickPeer(key string)(PeerGetter,bool){
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