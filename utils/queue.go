package utils

import (
	"unsafe"
	"time"
	"sync/atomic"
)

type TQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
	len  int64
}

func NewQueue() (*TQueue) {
	nn := &tNode{}
	return &TQueue{
		head:unsafe.Pointer(nn),
		tail:unsafe.Pointer(nn),
	}
}

func (this *TQueue)Offer(_val interface{})(int64) {
	node := &tNode{val:_val}
	for {
		tail := this.tail
		next := (*tNode)(tail).next
		if tail == this.tail {
			if nil == next {
				if atomic.CompareAndSwapPointer(&(*tNode)(this.tail).next, next, unsafe.Pointer(node)) {
					atomic.CompareAndSwapPointer(&this.tail, tail, unsafe.Pointer(node))
					atomic.AddInt64(&this.len, 1)
					break
				}
			} else {
				atomic.CompareAndSwapPointer(&this.tail, tail, next)
			}
		}
	}
	return this.Len()
}

func (this *TQueue)Poll(_timeout time.Duration) (val interface{}, ok bool) {
	val, ok = this.doPoll()
	for !ok && _timeout > 0 {
		time.Sleep(time.Millisecond)
		_timeout -= time.Millisecond
		val, ok = this.doPoll()
	}
	return
}

func (this *TQueue)doPoll() (interface{}, bool) {
	for {
		head := this.head
		tail := this.tail
		next := (*tNode)(head).next
		if head == this.head {
			if head == tail {
				if nil == next {
					return nil, false
				}
				atomic.CompareAndSwapPointer(&this.tail, tail, next)
			} else {
				val := (*tNode)(next).val
				if atomic.CompareAndSwapPointer(&this.head, head, next) {
					atomic.AddInt64(&this.len, -1)
					return val, true
				}
			}
		}
	}
}

func (this *TQueue)IsEmpty() (bool) {
	return this.Len() == 0
}

func (this *TQueue)Len() (int64) {
	return atomic.LoadInt64(&this.len)
}
