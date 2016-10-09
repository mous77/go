package utils

type TBufPool struct {
	TStack
	bufSize int
}

func NewBufPool(_pool_size, _buf_size int) (pool *TBufPool) {
	println("NewBufPool(poolSize:", _pool_size, ", bufSize:", _buf_size, ")")

	pool = &TBufPool{bufSize:_buf_size}
	for i := 0; i < _pool_size; i++ {
		pool.Push(NewBufObj(_buf_size))
	}
	return
}

func (this *TBufPool)Pop() (buf *TBufObj) {
	if obj, ok := this.TStack.Pop(); ok {
		buf = obj.(*TBufObj)
	} else {
		buf = NewBufObj(this.bufSize)
	}
	return
}

func (this *TBufPool)Push(_buf *TBufObj) {
	if nil != _buf {
		_buf.Clear()
		this.TStack.Push(_buf)
	}
}