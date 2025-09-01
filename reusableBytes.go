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

// NewReusableBytes 获取一个缓冲区
func NewReusableBytes(size int) *ReusableBytes {
	return &ReusableBytes{
		buffer: make([]byte, size),
		anchor: -1,
	}
}

// WriteString 往缓冲区中写入string
func (rb *ReusableBytes) WriteString(s string) int {
	if len(s) == 0 {
		return 0
	}
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

// Len 获取当前缓冲区的长度
func (rb *ReusableBytes) Len() int {
	return rb.cursor
}

// Reset 重置缓冲区
func (rb *ReusableBytes) Reset() {
	rb.cursor = 0
	rb.anchor = -1
}

// Anchor 设置锚点
func (rb *ReusableBytes) Anchor() {
	rb.anchor = rb.cursor
}

// Unanchor 取消锚点
func (rb *ReusableBytes) Unanchor() {
	rb.anchor = -1
}

// StringFromAnchor 获取从锚点下标开始的string
func (rb *ReusableBytes) StringFromAnchor() string {
	if rb.anchor == -1 || rb.cursor <= rb.anchor {
		return ""
	}
	return unsafe.String(&rb.buffer[rb.anchor], rb.cursor-rb.anchor)
}

// String 获取由缓存转化的string
func (rb *ReusableBytes) String() string {
	if rb.cursor == 0 || rb.buffer == nil {
		return ""
	}
	return unsafe.String(&rb.buffer[0], rb.cursor)
}

// Bytes 获取由缓冲区转化的byte切片
func (rb *ReusableBytes) Bytes() []byte {
	if rb.cursor == 0 || rb.buffer == nil {
		return nil
	}
	return rb.buffer[:rb.cursor]
}

// UnsafeBuffer 返回底层缓冲区的起始指针
func (rb *ReusableBytes) UnsafeBuffer() unsafe.Pointer {
	if rb.buffer == nil {
		return nil
	}
	return unsafe.Pointer(&rb.buffer[0])
}

// Grow 确保缓冲区至少还能再容纳 n 个字节
func (rb *ReusableBytes) Grow(n int) {
	if n < 0 {
		panic("reusablebytes: Grow with negative count")
	}
	needed := rb.cursor + n
	if needed <= cap(rb.buffer) {
		return
	}
	newCap := cap(rb.buffer) * 2
	if newCap < needed {
		newCap = needed
	}
	newBuf := make([]byte, newCap)
	copy(newBuf, rb.buffer[:rb.cursor])
	rb.buffer = newBuf
}

// Resize 将缓冲区调整到指定大小
// - 如果 size 大于当前容量，则重新分配并扩容
// - 如果 size 小于等于容量，则仅调整 cursor，相当于截断
func (rb *ReusableBytes) Resize(size int) {
	if size < 0 {
		panic("reusablebytes: Resize with negative size")
	}

	if size > cap(rb.buffer) {
		newBuf := make([]byte, size)
		if rb.cursor > 0 {
			// 保留现有数据
			copy(newBuf, rb.buffer[:rb.cursor])
		}
		rb.buffer = newBuf
	}

	rb.cursor = size
	if rb.cursor > len(rb.buffer) {
		// 防止逻辑长度超过底层切片长度
		rb.buffer = rb.buffer[:rb.cursor]
	} else {
		rb.buffer = rb.buffer[:cap(rb.buffer)]
	}
}
