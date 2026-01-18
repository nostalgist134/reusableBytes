package reusablebytes

import (
	"unsafe"
)

// ReusableBytes 可复用的内存缓存区，提供零拷贝的到字符串/字节流的转化方法
type ReusableBytes struct {
	buffer []byte
	cursor int
	anchor int
}

// Lazy 延迟分配结构，这个结构是为了用来解决如果ReusableBytes底层内存改变后，
// 原先分配的字符串/字节流与现有的底层内存不在同一个切片的问题（虽然这个问题实际
// 并不影响库的使用，但是总是感觉不太好），有了这个结构就能实现在
// ReusableBytes中全部写完，再转化字符串的功能
// 注意：不推荐将Lazy结构转为string或者字节流后，还调用rb.Write方法，虽然不会
// 引发错误，但是这个结构提供的功能会失效
type Lazy struct {
	rb    *ReusableBytes
	start int
	end   int
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
	copy(rb.buffer[rb.cursor:], s)
	rb.cursor += len(s)
	return len(s)
}

func (rb *ReusableBytes) WriteBytes(p []byte) int {
	if len(p) == 0 {
		return 0
	}
	needed := rb.cursor + len(p)
	if needed > len(rb.buffer) {
		newCap := len(rb.buffer) * 2
		if newCap < needed {
			newCap = needed
		}
		newBuf := make([]byte, newCap)
		copy(newBuf, rb.buffer)
		rb.buffer = newBuf
	}
	copy(rb.buffer[rb.cursor:], p)
	rb.cursor += len(p)
	return len(p)
}

func (rb *ReusableBytes) Write(p []byte) (n int, err error) {
	return rb.WriteBytes(p), nil
}

// Len 获取当前缓冲区的长度（即游标的位置）
func (rb *ReusableBytes) Len() int {
	return rb.cursor
}

// Cap 获取当前缓冲区在不重新分配的情况下的最大长度（底层切片的cap）
func (rb *ReusableBytes) Cap() int {
	return cap(rb.buffer)
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

// UnsafeBuffer 返回底层缓冲区的起始指针，注意：在扩容后（比如说写入、Grow、Resize）
// 应该尽量避免原先获取的UnsafeBuffer指针，因为此时底层缓冲区已经改变，虽然不会报错
// 但是往其中写入的数据不会同步到扩容后的对象
func (rb *ReusableBytes) UnsafeBuffer() unsafe.Pointer {
	if rb.buffer == nil {
		return nil
	}
	return unsafe.Pointer(&rb.buffer[0])
}

// Resize 调整游标或者缓冲区大小，当调整的size大于底层切片的cap时，切片扩容
// 否则只调整游标，不重新分配
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
}

// Lazy 将ReusableBytes结构转化为Lazy结构，返回一个字面值，因为要尽量避免堆内存分配
func (rb *ReusableBytes) Lazy() Lazy {
	return Lazy{
		rb:    rb,
		start: 0,
		end:   rb.cursor,
	}
}

// LazyFromAnchor 从锚点开始转化
func (rb *ReusableBytes) LazyFromAnchor() Lazy {
	return Lazy{
		rb:    rb,
		start: rb.anchor,
		end:   rb.cursor,
	}
}

// String 将Lazy转为string
func (l Lazy) String() string {
	if l.start < 0 || l.start > l.rb.cursor || l.end > l.rb.cursor || l.start == l.end {
		return ""
	}
	return unsafe.String(&l.rb.buffer[l.start], l.end-l.start)
}

// Bytes 将Lazy转为字节流
func (l Lazy) Bytes() []byte {
	if l.start < 0 || l.start > l.rb.cursor || l.end > l.rb.cursor || l.start == l.end {
		return []byte{}
	}
	return l.rb.buffer[l.start:l.end]
}

func (l Lazy) Len() int {
	return l.end - l.start
}
