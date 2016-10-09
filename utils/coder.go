package utils

import (
	"strings"
	"strconv"
	"encoding/hex"
)

/**
 * 解码  7D02  => 7E,  7D01  => 7D
 * @param _prex 前缀
 * @param _maps byte[][2] 映射表 [ {src, dst} ]
 */
func Decode(_src []byte, _prex byte, _map *[][2]byte, _dst *[]byte) {
	m := *_map
	dst := *_dst

	mc := len(m)
	cnt := len(_src)
	dst = append(dst, _src[0])
	for i := 1; i < cnt - 1; i++ {
		if _prex != _src[i] {
			dst = append(dst, _src[i])
		} else {
			for j := 0; j < mc; j++ {
				if _src[i + 1] == m[j][1] {
					dst = append(dst, m[j][0])
					break
				}
			}
			i++
		}
	}
	dst = append(dst, _src[cnt - 1])

	*_dst = dst
}

/**
 * 转义 0x7E <-----> 0x7D 0x02,
 *     0x7D <-----> 0x7D 0x01，
 *
 *   头尾不动
 * @param _data 数据
 * @param _prex 前缀
 * @param _maps 映射列表
 */
func Encode(_src []byte, _prex byte, _map *[][2]byte, _dst *[]byte) {
	m := *_map
	cnt := len(_src)
	mc := len(m)
	*_dst = append(*_dst, _src[0])
	for i := 1; i < cnt - 1; i++ {
		found := false
		for j := 0; j < mc; j++ {
			if _src[i] == m[j][0] {
				*_dst = append(*_dst, _prex, m[j][1])
				found = true
				break
			}
		}
		if !found {
			*_dst = append(*_dst, _src[i])
		}
	}
	*_dst = append(*_dst, _src[cnt - 1])
}

func TestCode() {
	_0XMAPS := [][2]byte{{0x7d, 0x01}, {0x7d, 0x02}}
	hh := "7E80010005013789300274000200020102007D027E"
	println("src=" + hh)
	buf := make([]byte, 64)

	src, _ := hex.DecodeString(hh)
	dst := buf[0:0]

	Decode(src, 0x7D, &_0XMAPS, &dst)
	println("dst=" + hex.EncodeToString(dst))
}

func StrToByte(_s string) (byte) {
	v, _ := strconv.Atoi(_s)
	return byte(v)
}

func HexToInt(_s string) (int) {
	v, _ := strconv.ParseInt(_s, 16, 0)
	return int(v)
}

func IndexOf(_buf *[]byte, _offset, _len int, _key byte) (int) {
	idx := _offset
	slice := *_buf
	for ; idx < _len; {
		if slice[idx] == _key {
			break
		} else {
			idx++
		}
	}
	return idx
}

func ExtractAddr(_ip_port string) (string, int, bool) {
	ary := strings.Split(_ip_port, ":")
	host := ary[0]
	if port, err := strconv.Atoi(ary[1]); nil != err {
		return host, port, true
	} else {
		return "", 0, false
	}
}

const c_BKRD_SEED uint32 = 131 // 31 131 1313 13131 131313 etc...

func BKRDHash(_src string) uint32 {
	var h uint32
	for _, c := range _src {
		h = h * c_BKRD_SEED + uint32(c)
	}
	return h
}