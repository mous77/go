package net

import(
	"github.com/mous77/go/utils"
)

type TOption struct {
	RcvQLen  int
	SndQLen  int
	BufPool  *utils.TBufPool
	PoolSize int
	BufSize  int
	Timeout  int// 5分钟没有收到数据认为掉线
}

func NewOption() (*TOption) {
	return &TOption{
		RcvQLen:8,
		SndQLen:8,
		PoolSize:8192,
		BufSize:1024,
		Timeout:300}
}

func (_self *TOption)copyFrom(_src *TOption) {
	if nil != _src {
		if _src.RcvQLen > 0 {
			_self.RcvQLen = _src.RcvQLen
		}

		if _src.SndQLen > 0 {
			_self.SndQLen = _src.SndQLen
		}

		if _src.PoolSize > 0 {
			_self.PoolSize = _src.PoolSize
		}

		if _src.BufSize > 0 {
			_self.BufSize = _src.BufSize
		}

		if _src.BufPool != nil{
			_self.BufPool = _src.BufPool
		}
	}
}

func (_self *TOption)GetBufPool()(*utils.TBufPool){
	if nil==_self.BufPool{
		return utils.NewBufPool(_self.PoolSize, _self.BufSize)
	}else {
		return _self.BufPool
	}
}
