package reusablebytes

import (
	"unsafe"
	_ "unsafe"
)

type ReusableBytes struct {
	buffer []byte
	cursor int
	anchor int
}

//go:linkname memmove runtime.memmove
func memmove(to, from unsafe.Pointer, n uintptr)

func copyString(dst []byte, offset int, s string) {
	memmove(
		unsafe.Pointer(&dst[offset]),
		unsafe.Pointer(unsafe.StringData(s)),
		uintptr(len(s)),
	)
}

func NewReusableBytes(size int) *ReusableBytes {
	return &ReusableBytes{
		buffer: make([]byte, size),
		anchor: -1,
	}
}

func (rb *ReusableBytes) WriteString(s string) int {
	needed := rb.cursor + len(s)
	if needed > len(rb.buffer) {
		newCap := len(rb.buffer) * 2
		if newCap < needed {
			newCap = needed
		}
		newBuf := make([]byte, newCap)
		copy(newBuf, rb.buffer)
		rb.buffer = newBuf
	}
	copyString(rb.buffer, rb.cursor, s)
	rb.cursor += len(s)
	return len(s)
}

func (rb *ReusableBytes) Reset() {
	rb.cursor = 0
	rb.anchor = -1
}

func (rb *ReusableBytes) Anchor() {
	rb.anchor = rb.cursor
}

func (rb *ReusableBytes) Unanchor() {
	rb.anchor = -1
}

func (rb *ReusableBytes) StringFromAnchor() string {
	if rb.anchor == -1 || rb.cursor <= rb.anchor {
		return ""
	}
	return unsafe.String(&rb.buffer[rb.anchor], rb.cursor-rb.anchor)
}

func (rb *ReusableBytes) String() string {
	if rb.cursor == 0 {
		return ""
	}
	return unsafe.String(&rb.buffer[0], rb.cursor)
}
