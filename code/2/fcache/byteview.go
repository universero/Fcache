package fcache

// A ByteView holds an immutable view of bytes
// It's one of the most important structure in the fcache
// b is only read
type ByteView struct {
	b []byte
}

// Len returns the view's length
// It's realise the interface
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice return a new slice which is deeplyEqual with the field b
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String return the byte slice in string type
func (v ByteView) String() string {
	return string(v.b)
}

// deeply copy the byte slice
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
