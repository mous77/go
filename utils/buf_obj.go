package utils

import (
	"encoding/hex"
)

type TBufObj struct {
	_data []byte
	slice []byte
}

func NewBufObj(_size int) (*TBufObj) {
	data := make([]byte, _size)
	return &TBufObj{_data:data, slice:data[0:0]}
}

func (this *TBufObj)Write(_buf ...byte) {
	this.slice = append(this.slice, _buf...)
}

func (this *TBufObj)WriteBuf(_buf *TBufObj){
	this.slice = append(this.slice, _buf.slice...)
}

func (this *TBufObj)WriteString(_str string){
	this.Write([]byte(_str)...)
}

func (this *TBufObj)Clear() {
	this.slice = this._data[0:0]
}

func (this *TBufObj)Size() (int) {
	return len(this.slice)
}

func (this *TBufObj)ByteAt(_index int)(byte){
	return this.slice[_index]
}

func (this *TBufObj)GetMem(_size int) ([]byte) {
	return this._data[0:_size]
}

func (this *TBufObj)Read(_size int, _dst *[]byte){
	if _size > this.Size(){
		_size = this.Size()
	}
	if nil!=_dst{
		*_dst = append(*_dst, this.slice[0:_size]...)
	}
	this.slice = this.slice[_size:]
}

func (this *TBufObj)IndexOf(_who byte, _offset int) (int) {
	for i := _offset; i < len(this.slice); i++ {
		if this.slice[i] == _who {
			return i
		}
	}
	return -1
}

func (this *TBufObj)ToHex() (string) {
	return hex.EncodeToString(this.slice)
}

func (this *TBufObj)Slice()([]byte){
	return this.slice
}

func (this *TBufObj)Data()([]byte){
	return this._data
}

func (this *TBufObj)String()(string){
	return string(this.slice)
}