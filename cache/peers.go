package cache

// this picker will return a getter function, directly fetch data from other node
type PeerPicker interface{
	PickerPeer(key string)(peer PeerGetter, ok bool)
}

// function to return data from another node
type PeerGetter interface{
	Get(group string, key string)([]byte,error)
}