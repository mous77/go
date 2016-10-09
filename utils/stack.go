package utils

import (
	"sync/atomic"
	"unsafe"
	"time"
)

type tNode struct {
	val  interface{}
	next unsafe.Pointer // *tNode
}

type TStack struct {
	top unsafe.Pointer // *tNode
	len int32
}

func NewStack() (stack *TStack) {
	return &TStack{}
}

/*
	返回堆栈高度
 */
func (this *TStack)Len() int32 {
	return atomic.LoadInt32(&this.len)
}

/*
	数据入栈
 */
func (this *TStack)Push(_val interface{}) {
	if nil != _val {
		node := &tNode{val:_val}
		for {
			node.next = this.top
			if atomic.CompareAndSwapPointer(&this.top, node.next, unsafe.Pointer(node)) {
				atomic.AddInt32(&this.len, 1)
				break
			}
		}
	}
}

func (this *TStack)Pop() (interface{}, bool) {
	for {
		top := this.top
		if nil == top {
			return nil, false
		} else {
			node := (*tNode)(top)
			if atomic.CompareAndSwapPointer(&this.top, top, node.next) {
				atomic.AddInt32(&this.len, -1)
				return node.val, true
			}
		}
	}
	return nil, false
}

func TestStack() {
	println("testStack")
	stack := NewStack()

	type node struct {
		val int
	}

	println("setup")
	for i := 0; i < 1024; i++ {
		stack.Push(&node{val:i})
	}

	println("go 1")
	go func() {
		for {
			if val, ok := stack.Pop(); ok {
				nd := val.(*node)
				println("1.pop= ", nd.val)
				stack.Push(nd)
			}
			println("-----")
			time.Sleep(time.Millisecond)
		}
	}()

	println("go 2")
	go func() {
		for {
			if val, ok := stack.Pop(); ok {
				nd := val.(*node)
				println("2.pop= ", nd.val)
				stack.Push(nd)
			}
			println("========")
			time.Sleep(time.Millisecond)
		}
	}()

	tm := time.NewTimer(time.Second * 5)
	println("start timer")
	for {
		select {
		case <-tm.C:
			println("time")
			tm.Reset(time.Second * 5)
		}
	}
}