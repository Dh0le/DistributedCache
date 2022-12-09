package cache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string)([]byte,error){
		return []byte(key),nil
	})
	expect := []byte("key")
	if v,_ := f.Get("key");!reflect.DeepEqual(v,expect){
		t.Fatal("call back failed")
	}
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	gee := NewGroup("scores",2<<10,GetterFunc(func(key string)([]byte,error){
		log.Print("[SlowDB] search key",key)
		if v,ok := db[key];ok{
			if _,ok := loadCounts[key];!ok{
				loadCounts[key] = 0
			}
			loadCounts[key]++
			return []byte(v),nil
		}
		return nil, fmt.Errorf("%s not exist",key)
	}))

	for k,v := range db{
		if view,err := gee.Get(k);err != nil || view.String() != v{
			t.Fatal("Failed to get value")
		}
		if _, err := gee.Get(k);err != nil|| loadCounts[k] > 1{
			t.Fatalf("cache %s missed",k)
		}
		if view,err := gee.Get("unknown");err == nil{
			t.Fatalf("the value of known should be empty, got %s",view)
		}
	}
}

func TestGroup(t *testing.T) {
	groupName := "scores"
	NewGroup(groupName,2<<10,GetterFunc(func(key string)(bytes []byte,err error){
		return 
	}))

	if group := GetGroup(groupName);group == nil||group.name != groupName{
		t.Fatalf("Group not found")
	}

	if group := GetGroup("1111");group != nil{
		t.Fatalf("Expect nil but got %s",group.name)
	}


}


