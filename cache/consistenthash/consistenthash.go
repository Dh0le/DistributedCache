package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)


type Hash func(data []byte)uint32
// create a hash interface that allows user to inject customized hash function

type Map struct{
	hash Hash // customized hash function injection
	replicas int // how many virtual node you want to have
	keys []int // hash ring
	hashMap map[int]string // mapping relation between key on hash ring and real hash server
}

// consistent hashing constructor
func New(replicas int, fn Hash)*Map{
	m := &Map{
		replicas: replicas,
		hash: fn,
		hashMap: make(map[int]string),
	}
	// default consistent hashing function
	if m.hash == nil{
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map)Add(keys ...string){
	// for every key, we generate number of virtual nodes and place them into hash ring
	for _,key := range keys{
		for i:= 0;i < m.replicas;i++{
			hash := int(m.hash([]byte(strconv.Itoa(i)+key)))
			m.keys = append(m.keys,hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// given a key, return the server it stored in
func(m *Map)Get(key string)string{
	// null check for hash ring
	if len(m.keys) == 0{
		return ""
	}

	// for current hash key get its value on hash ring
	hash := int(m.hash([]byte(key)))
	// get search clockwise to find the first idx
	idx := sort.Search(len(m.keys),func(i int)bool{
		return m.keys[i] >= hash
	})
	// get its mapping server and return
	return m.hashMap[m.keys[idx%len(m.keys)]]
}

