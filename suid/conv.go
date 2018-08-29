package suid

import (
	"bytes"
)

const radix = 62

var digitalAry62 = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func digTo62(_val int64, _digs byte, _sb *bytes.Buffer) {
	hi := int64(1) << (_digs * 4)
	i := hi | (_val & (hi - 1))

	negative := i < 0
	if !negative {
		i = -i
	}

	skip := true
	for i <= -radix {
		if skip {
			skip = false
		} else {
			offset := -(i % radix)
			_sb.WriteByte(digitalAry62[int(offset)])
		}
		i = i / radix
	}
	_sb.WriteByte(digitalAry62[int(-i)])

	if (negative) {
		_sb.WriteByte('-')
	}
}

func suidToShortS(data []byte) string {
	// [16]byte
	buf := make([]byte, 22)
	sb := bytes.NewBuffer(buf)
	sb.Reset()

	var msb int64
	for i := 0; i < 8; i++ {
		msb = msb<<8 | int64(data[i])
	}

	var lsb int64
	for i := 8; i < 16; i++ {
		lsb = lsb<<8 | int64(data[i])
	}

	digTo62(msb>>12, 8, sb)
	digTo62(msb>>16, 4, sb)
	digTo62(msb, 4, sb)
	digTo62(lsb>>48, 4, sb)
	digTo62(lsb, 12, sb)

	return sb.String()
}
