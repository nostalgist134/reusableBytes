package reusablebytes

import (
	"sync"
)

type BytesPool struct {
	pool      []*ReusableBytes
	usableInd []int32
	available int32
	mu        sync.Mutex
	cond      *sync.Cond
	maxSize   int32
	initSize  int32
}

// Init 初始化池，initSize 是初始容量，maxSize 是最大容量
func (p *BytesPool) Init(initSize, maxSize int32, eachBytesLen int) {
	p.pool = make([]*ReusableBytes, initSize)
	p.usableInd = make([]int32, initSize)
	for i := int32(0); i < initSize; i++ {
		p.pool[i] = NewReusableBytes(eachBytesLen)
		p.usableInd[i] = i
	}
	p.available = initSize
	p.maxSize = maxSize
	p.initSize = initSize
	p.cond = sync.NewCond(&p.mu)
}

// Get 获取一个可用的 ReusableBytes 对象，自动 Reset
func (p *BytesPool) Get() (*ReusableBytes, int32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		if p.available > 0 {
			p.available--
			id := p.usableInd[p.available]
			rb := p.pool[id]
			rb.Reset()
			return rb, id
		}

		// 可扩容
		if int32(len(p.pool)) < p.maxSize {
			newId := int32(len(p.pool))
			newBuf := NewReusableBytes(128)
			p.pool = append(p.pool, newBuf)
			p.usableInd = append(p.usableInd, newId)
			// 注意新加对象直接返回，不入栈，因为已经被“占用”
			return newBuf, newId
		}

		// 无可用，等待
		p.cond.Wait()
	}
}

// Put 归还 ReusableBytes 对象，传入对象id
func (p *BytesPool) Put(id int32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if id < 0 || id >= int32(len(p.pool)) {
		return
	}

	// 防止超过栈容量，通常不会超过，做个安全判断
	if p.available < int32(len(p.usableInd)) {
		p.usableInd[p.available] = id
		p.available++
		p.cond.Signal()
	}
}
