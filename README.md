# DistributedCache

Distributed Concurrent Cache based on Golang

## Overall

This is a distributed concurrent Key-Value cache based on Go and Gin

### Structure diagram

![avatar](/Distributed%20Cache%20diagram.jpeg)

### Cache Algorithm:

The cache is implementted with LRU with Go std lib container double linked list.

### Cache Type: Cache Through

I think cache through is somehow more conventient.
Developers do not need to interact with the database directly and worried about consistentency problem.

### Usage

Compile the project

```
go build -o yourCache
```

start a single cache server (you should create your getter function first)

```
./ yourCache -port=8888 -api=1
```

config distributed system

```
// here is all address of your nodes using they port number as index
// if you wanto test it in local env
addrMap := map[int]string{
		8001:"http://localhost:8001",
		8002:"http://localhost:8002",
		8003:"http://localhost:8003",
}

// create a array contains all peer nodes address
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

if api{
		// since we use gin as our sever, we need to use "go" to start a new thread
		// if we dont use go here, the thread will stuck and will not proceed to create local cache server
		go cache.StartAPIServer(apiAddr,":9999",cacheGroup)
}

// start Cache server
cache.StartCacheServer(addrMap[port],":"+ strconv.Itoa(port),addrs,cacheGroup)
```

How to create your Getter function
Mysql for example

```
getterFn := cache.GetterFunc(func(key string) ([]byte, error) {
	// create connection to you database in here we use mysql as example
  // in real env, connection should be placed in its own module
  DB, _ := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/test")
  DB.SetConnMaxLifetime(100)
  DB.SetMaxIdleConns(10)
  if err := DB.Ping(); err != nil {
    fmt.Println("open database fail")
    return nil,err
  }
  fmt.Println("connnect success")
  // start a query
  var user User
  DB.QueryRow("select * from user where id=1").Scan(user.age, user.id, user.name, user.phone, user.sex)
  return []byte(user),nil
})
```
