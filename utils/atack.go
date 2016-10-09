package utils

import (
	"sync/atomic"
	"unsafe"
	"time"
)

type Atack struct {
	data []unsafe.Pointer
	top  int32
	cap  int32
}

func NewAtack(_size int32) (*Atack) {
	return &Atack{data:make([]unsafe.Pointer, _size), cap:_size}
}

/*
	数据入栈
 */
func (s *Atack)Push(_val interface{}) {
	if nil != _val {
		cur := &_val
		for s.top < s.cap {
			if atomic.CompareAndSwapPointer(&s.data[s.top], nil, unsafe.Pointer(cur)) {
				atomic.AddInt32(&s.top, 1)
				break
			}
		}
	}
}

func (s *Atack)Pop() (interface{}, bool) {
	for {
		top := s.top
		if top < 1 {
			return nil, false
		} else {
			val := s.data[top-1]
			if nil==val {
				return nil, false
			}else {
				if atomic.CompareAndSwapPointer(&s.data[top - 1], val, nil) {
					atomic.AddInt32(&s.top, -1)
					return *(*interface{})(val), true
				}
			}
		}
	}
	return nil, false
}

func TestAtack() {
	println("testStack")
	s := NewAtack(8)

	type node struct {
		val int
	}

	println("setup")
	for i := 0; i < 4; i++ {
		s.Push(&node{val:i*1000})
	}

	/*
	println("go 1")
	go func() {
		for {
			if val, ok := s.Pop(); ok {
				nd := val.(*node)
				println("1.pop= ", nd.val)
				s.Push(val)
			}
			println("-----")
			time.Sleep(time.Second)
		}
	}()
	//*/

	println("go 2")
	go func() {
		for {
			if val, ok := s.Pop(); ok {
				nd := val.(*node)
				println("2.pop= ", nd.val)
				s.Push(val)
			}
			println("========")
			time.Sleep(time.Second)
		}
	}()

	tm := time.NewTimer(time.Second)
	println("start timer")
	for {
		select {
		case <-tm.C:
			println("s:", s.top, s.data[0], s.data[1], s.data[2], s.data[3])
			tm.Reset(time.Second )
		}
	}
}