package utils

import (
	"sync/atomic"
)

type tRingNode struct {
	prev *tRingNode
	next *tRingNode
	val  interface{}
}

func (this *tRingNode)clean() {
	this.next = nil
	this.prev = nil
	this.val = nil
}

type TRingList struct {
	head *tRingNode
	tail *tRingNode
	len  int32
}

type IRingIter interface {
	HasMore() (bool)
	Next() (interface{})
}

type tRingIter struct {
	owner *TRingList
	node  *tRingNode
}

func (this *tRingIter)HasMore()(bool) {
	if nil == this.node {
		this.node = this.owner.head
	}

	return nil != this.node && nil != this.node.val
}

func (this *tRingIter)Next() (interface{}) {
	val := this.node.val
	this.node = this.node.next
	return val
}

func NewRingList() (*TRingList) {
	return &TRingList{}
}

func (this *TRingList)NewIter() (IRingIter) {
	return &tRingIter{owner:this, node:nil}
}

func (this *TRingList)Len() (int32) {
	return atomic.LoadInt32(&this.len)
}

func (this *TRingList)Add(_val interface{}) {
	node := &tRingNode{prev:this.tail, next:nil, val:_val}
	this.tail = node
	if nil == this.head {
		this.head = node
	}
	atomic.AddInt32(&this.len, 1)
}

func (this *TRingList)Del(_val interface{}) {
	node := this.find(_val)
	if nil != node {
		prev := node.prev
		next := node.next

		if nil != prev {
			prev.next = next
		}

		if nil != next {
			next.prev = prev
		}

		if this.head == node {
			this.head = next
		}

		if this.tail == node {
			this.tail = prev
		}

		node.clean()
		atomic.AddInt32(&this.len, -1)
	}
}

func (this *TRingList)find(_val interface{}) (*tRingNode) {
	tmp := this.head
	for nil != tmp {
		if tmp.val == _val {
			return tmp
		} else {
			tmp = tmp.next
		}
	}
	return nil
}
