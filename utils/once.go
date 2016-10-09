package utils

import(
	"sync/atomic"
)

type TOnce struct {
	done int32
}

func (this *TOnce) Do(f func()) {
	if atomic.LoadInt32(&this.done) == 1 {
		return
	}

	if atomic.CompareAndSwapInt32(&this.done, 0, 1) {
		f()
	}
}
