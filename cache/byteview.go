package cache

// So the ByteView is what we store in cache as value
// This is a read only struct, since we don't want the value in cache be modifed from outside
type ByteView struct{
	// we store value in bytes arr
	b []byte
}

// this struct must implement Len method to be Value interface
func(v ByteView)Len()int{
	return len(v.b)
}

// return a clone
func(v ByteView)ByteSlice()[]byte{
	return cloneByte(v.b)
}
// return a string
func(v ByteView)String()string{
	return string(v.b)
}

// internal clone method
func cloneByte(b[]byte)[]byte{
	c := make([]byte,len(b))
	copy(c,b);
	return c
}