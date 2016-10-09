package net

import (
	"github.com/mous77/go/utils"
	"strconv"
	"strings"
	"fmt"
)

type TunnelType byte

const (
	TUN_TCP TunnelType = 'T'
	TUN_UDP TunnelType = 'U'
)

/** 0..14
 *                     0        12       3456       78           9012       34
 * 会话地址 bytes[15] = Tun[1] + SEQ[2] + SvrIP[4] + SvrPort[2] + CliIP[4] + CliPort[2]
 * [0]    		: Tunnel	= value[0:1)
 * [1..2] 		: SEQ		= value[1:3)
 * [3..6]		: SvrIP		= value[3:7)
 * [7..8]    	: SvrPort	= value[7:9)
 * [9..12]		: CliIP		= value[9:13)
 * [13..14] 	: CliPort	= value[13:15]
 *
 */
type TNetAddr struct {
	val []byte
}

func NewAddr() (*TNetAddr) {
	return &TNetAddr{val:make([]byte, 15)}
}

func WrapAddr(_src []byte) (*TNetAddr) {
	slice := _src[:15]
	return &TNetAddr{val:slice}
}

func MakeAddrByString(_src *string) (*TNetAddr) {
	slice := SessionKeyToBytes(_src)
	return &TNetAddr{val:slice}
}

func (this *TNetAddr)SetTunnel(_tunnel TunnelType) {
	this.val[0] = byte(_tunnel)
}

func (this *TNetAddr)GetTunnel() (TunnelType) {
	return TunnelType(this.val[0])
}

func (this *TNetAddr)GetValue() ([]byte) {
	return this.val
}

func (this *TNetAddr)String() (string) {
	return BytesToSessionKey(this.val)
}

func (this *TNetAddr)CopyTo(_dst []byte, _offset int) {
	src := this.val[0:15]
	dst := _dst[_offset:_offset + 15]
	copy(dst, src)
}

func (this *TNetAddr)Clear() {
	for i := 0; i < 15; i++ {
		this.val[i] = 0
	}
}

func (this *TNetAddr)CopyFrom(_src *TNetAddr) {
	_src.CopyTo(this.val, 0)
}

func (this *TNetAddr)GetLocalPort() (int) {
	return int(this.val[7] >> 8 | this.val[8])
}

/**
 * 获取服务端地址
 * @return 服务端地址
 */
func (this *TNetAddr)GetLocal() ([]byte) {
	return this.val[3:9]
}

/**
 * 获取远端地址
 * @return 远端地址
 */
func (this *TNetAddr)GetRemote() ([]byte) {
	return this.val[9:15]
}

/**
 * 地址化为字符数组	   0 1    2  3   4 5 6     7  8  9   10  11
 * @param _session 地址 T0000-192.168.5.7:6000-202.98.232.22:4097
 * @return 字符数据
 */
func sessionKeyToAry(_key *string) ([]string) {
	ary := make([]string, 12)

	var sx, ex int

	buf := []byte(*_key)
	len := len(buf)

	ary[0] = string(buf[0:1])
	ary[1] = string(buf[1:5])

	ex = utils.IndexOf(&buf, 6, len, '.');
	ary[2] = string(buf[6:ex]); sx = ex + 1
	ex = utils.IndexOf(&buf, sx, len, '.');
	ary[3] = string(buf[sx:ex]); sx = ex + 1
	ex = utils.IndexOf(&buf, sx, len, '.')
	ary[4] = string(buf[sx:ex]); sx = ex + 1
	ex = utils.IndexOf(&buf, sx, len, ':')
	ary[5] = string(buf[sx:ex]); sx = ex + 1

	ex = utils.IndexOf(&buf, sx, len, '/')
	ary[6] = string(buf[sx:ex]); sx = ex + 1

	ex = utils.IndexOf(&buf, sx, len, '.')
	ary[7] = string(buf[sx:ex]); sx = ex + 1
	ex = utils.IndexOf(&buf, sx, len, '.')
	ary[8] = string(buf[sx:ex]); sx = ex + 1
	ex = utils.IndexOf(&buf, sx, len, '.')
	ary[9] = string(buf[sx:ex]); sx = ex + 1
	ex = utils.IndexOf(&buf, sx, len, ':')
	ary[10] = string(buf[sx:ex])

	ary[11] = string(buf[ex + 1:])

	return ary
}

/**
 * 地址化为二进制表示        0 1    2   3  4 5 6     7   8  9  10  11
 * @param _session 会话地址 T:999-192.168.5.7:6000/202.98.232.22:4097
 * @return 字节数组 [ 1 + 2 + 6 + 6 ]
 */
func SessionKeyToBytes(_key *string) ([]byte) {
	buf := make([]byte, 15)
	tmp := sessionKeyToAry(_key)

	if strings.Compare("T", tmp[0]) == 0 {
		buf[0] = byte(TUN_TCP)
	} else {
		buf[0] = byte(TUN_UDP)
	}
	seq := utils.HexToInt(tmp[1])
	buf[1] = byte(seq >> 8)
	buf[2] = byte(seq)

	buf[3] = utils.StrToByte(tmp[2])
	buf[4] = utils.StrToByte(tmp[3])
	buf[5] = utils.StrToByte(tmp[4])
	buf[6] = utils.StrToByte(tmp[5])

	port, _ := strconv.Atoi(tmp[6])
	buf[7] = byte(port >> 8)
	buf[8] = byte(port)

	buf[9] = utils.StrToByte(tmp[7])
	buf[10] = utils.StrToByte(tmp[8])
	buf[11] = utils.StrToByte(tmp[9])
	buf[12] = utils.StrToByte(tmp[10])

	port, _ = strconv.Atoi(tmp[11])
	buf[13] = byte(port >> 8)
	buf[14] = byte(port)

	return buf
}

func makeWord(_b1, _b2 byte) (int) {
	return int(int(_b1) << 8 | int(_b2))
}

func makeIP(p []byte) (string) {
	return fmt.Sprintf("%d.%d.%d.%d", p[0], p[1], p[2], p[3])
}

func IpPortToStr(p []byte) (string) {
	return fmt.Sprintf("%d.%d.%d.%d:%d", p[0], p[1], p[2], p[3], makeWord(p[4], p[5]))
}

/**
 *                     0        12       3456       78           9012       34
 * 会话地址 bytes[15] = Tun[1] + SEQ[2] + SvrIP[4] + SvrPort[2] + CliIP[4] + CliPort[2]
 */
func BytesToSessionKey(p []byte) (string) {
	svr := IpPortToStr(p[3:9])
	cli := IpPortToStr(p[9:15])
	return fmt.Sprintf("%c%.4X-%s/%s", p[0], makeWord(p[1], p[2]), svr, cli)
}