package singleflight

import "sync"

type call struct{
	wg sync.WaitGroup
	val interface{}
	err error
}

type Group struct{
	mu sync.Mutex
	m map[string]*call
}

func (g *Group)Do(key string, fn func()(interface{},error))(interface{},error){
	// this lock makes sure map does not get concurrently modified
	g.mu.Lock()
	// delay intialization
	if g.m == nil{
		g.m = make(map[string]*call)
	}
	// if there is a ongoing call with same key
	if c,ok := g.m[key];ok{
		//(1)
		// wait until current ongoing call is finished, and return its value
		g.mu.Unlock()
		c.wg.Wait()
		return c.val,c.err
	}

	// if there is no ongoing call,create a ongoing call
	c := new(call)
	// locked befre it was called
	c.wg.Add(1)
	// add it into map
	g.m[key] = c
	g.mu.Unlock()

	// call fn to get value
	// so the trick is overhere, if our server is stuck with some request, before it actually returns value, no more same request will be sent to our server
	// extra request will be handled in (1)
	c.val, c.err = fn()
	c.wg.Done() // request is finished

	//update g.map
	g.mu.Lock()
	delete(g.m,key)
	g.mu.Unlock()
	//return value
	return c.val,c.err
}