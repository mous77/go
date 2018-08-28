/**
 * 2018-02-12, create by mous
 * 内存映射文件 读 写 封装
 */
package utils

import (
	"os"
	"time"
	"sync/atomic"
	"path"
	"syscall"
	"fmt"
	"bytes"
	"lib/log"
	"sync"
	"encoding/binary"
	"os/signal"
	"strings"
	"errors"
)

// 常量定义部分
//
const (
	rwFlagPos07  = 7  // 数据标识
	rwReadPos08  = 8  // 从头读取
	rwWritePos12 = 12 // 从尾写入
	rwDataPos16  = 16 // 数据所在
	rwFileProt   = syscall.PROT_READ | syscall.PROT_WRITE
	rwReadSize   = 4 * 1024 * 1024
)

var (
	rwKeyFlags = []byte("MAPS\r\n")
)

// 接口定义部分
//
type (
	// 公共定义部分
	ILineFileBase interface {
		Start()
		Stop()
		IsActive() bool
		String() string
	}

	// 数据读取接口
	ILineFileReader interface {
		ILineFileBase
		Pause()
		Resume()
	}

	// 数据写入接口
	ILineFileWriter interface {
		ILineFileBase
		Write(string)
	}

	// 数据读取部分
	//
	tLineFileBase struct {
		lg log.ILogger

		wgQuit *sync.WaitGroup
		file   *os.File
		fd     int
		fSize  int
		mbb    []byte
		fpath  string
		fname  string
		active int32
		queue  chan *string

		offsetR uint32
		offsetW uint32

		onStart func()
	}

	tLineFileReader struct {
		*tLineFileBase

		listener func(string)
		hasPause bool // 读取暂停标识
	}

	tLineFileWriter struct {
		*tLineFileBase
	}
)

// 公共部分
//
func (rw *tLineFileBase) clean() {
	defer func() {
		if v := recover(); nil != v {
			rw.lg.Error("error on clean %v", v)
		}
	}()

	rw.lg.Info("clean")

	if v := recover(); nil != v {
		rw.lg.Error("err %v \r\n ", v)
	}

	if nil != rw.mbb {
		syscall.Munmap(rw.mbb)
		rw.mbb = nil
	} else {
		rw.lg.Warn("mbb is null")
	}

	if nil != rw.file {
		rw.file.Close()
		rw.file = nil
	}
}

func (rw *tLineFileBase) Start() {
	if nil == rw.onStart {
		panic(errors.New("onStart is nil"))
	} else {
		rw.onStart()
	}
}

func (rw *tLineFileBase) Stop() {
	rw.lg.Info("Stop")
	atomic.StoreInt32(&rw.active, 0)
	rw.wgQuit.Wait()
	close(rw.queue)
}

func (rw *tLineFileBase) IsActive() (bool) {
	return atomic.LoadInt32(&rw.active) > 0
}

func (rw *tLineFileBase) String() string {
	if nil == rw.mbb {
		return fmt.Sprintf("[%s](正在打开...)", rw.fpath)
	} else {
		rPos := float32(rw.offsetR*100.0) / float32(rw.fSize)
		wPos := float32(rw.offsetW*100.0) / float32(rw.fSize)
		dVal := int64(rw.offsetW - rw.offsetR)
		return fmt.Sprintf("[%s][读:%.4f%%, 写:%.4f%%, 余:%s]", rw.fname, rPos, wPos, BytesToKMB(dVal))
	}
}

func (rw *tLineFileBase) mMap() (bool) {
	// func Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	// 参数fd为即将映射到进程空间的文件描述字，一般由open()返回，
	// 	同时，fd可以指定为-1，此时须指定flags参数中的MAP_ANON，表明进行的是匿名映射（不涉及具体的文件名，避免了文件的创建及打开，很显然只能用于具有亲缘关系的进程间通信）。
	//
	// len是映射到调用进程地址空间的字节数，它从被映射文件开头offset个字节开始算起。
	//
	// prot参数指定共享内存的访问权限。可取如下几个值的或：PROT_READ（可读）,PROT_WRITE（可写）,PROT_EXEC（可执行）,PROT_NONE（不可访问）。
	//
	// flags由以下几个常值指定：MAP_SHARED, MAP_PRIVATE, MAP_FIXED。
	// 	其中，MAP_SHARED,MAP_PRIVATE必选其一，而MAP_FIXED则不推荐使用。
	// 	如果指定为MAP_SHARED，则对映射的内存所做的修改同样影响到文件。
	// 	如果是MAP_PRIVATE，则对映射的内存所做的修改仅对该进程可见，对文件没有影响。
	//
	// offset参数一般设为0，表示从文件头开始映射。
	//
	// 参数addr指定文件应被映射到进程空间的起始地址，一般被指定一个空指针，此时选择起始地址的任务留给内核来完成。
	//
	// 函数的返回值为最后文件映射到进程空间的地址，进程可直接操作起始地址为该值的有效地址。
	if mbb, err := syscall.Mmap(rw.fd, 0, rw.fSize, rwFileProt, syscall.MAP_SHARED); nil != err {
		rw.lg.Error("error on Mmap(%s) %s", rw.fpath, err.Error())
		rw.clean()
		return false
	} else {
		rw.lg.Debug("mMap(%s) OK", rw.fpath)
		rw.mbb = mbb
		return true
	}
}

func (rw *tLineFileBase) loadRWOffset() {
	rw.offsetR = binary.BigEndian.Uint32(rw.mbb[rwReadPos08: rwReadPos08+4])
	rw.offsetW = binary.BigEndian.Uint32(rw.mbb[rwWritePos12: rwWritePos12+4])

	//rw.lg.Debug("loadRWOffset. R=%d, W=%d", rw.offsetR, rw.offsetW)
}

func (rw *tLineFileBase) saveRWOffset() {
	//rw.lg.Debug("saveRWOffset(R=%d, W=%d)", rw.offsetR, rw.offsetW)

	if rw.offsetR == rw.offsetW {
		rw.mbb[rwFlagPos07] = 0 // 读完了，没有了

		// 位置过半则调到 0
		if rw.offsetW > uint32(rw.fSize/2) {
			rw.offsetW = 0
			rw.offsetR = 0
		}
	} else {
		rw.mbb[rwFlagPos07] = 1
	}

	binary.BigEndian.PutUint32(rw.mbb[rwReadPos08: rwReadPos08+4], rw.offsetR)
	binary.BigEndian.PutUint32(rw.mbb[rwWritePos12: rwWritePos12+4], rw.offsetW)
}

func (rw *tLineFileBase) hasData() (bool) {
	return rw.mbb[rwFlagPos07] > 0
}

func hasFile(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func getName(fpath string) (string) {
	if idx := strings.LastIndex(fpath, "/"); idx > 0 {
		fpath = fpath[idx+1:]
	}
	return fpath
}

// 读部分
func NewLineFileReader(fpath string, listener func(string)) (ILineFileReader) {
	r := &tLineFileReader{
		tLineFileBase: &tLineFileBase{
			lg:     log.GetLogger("LReader"),
			wgQuit: &sync.WaitGroup{},
			fpath:  fpath,
			fname:  getName(fpath),
			active: 0,
			queue:  make(chan *string, 1024),
		},
		listener: listener,
	}
	r.onStart = r.doStartR
	return r
}

func (r *tLineFileReader) doStartR() {
	if atomic.CompareAndSwapInt32(&r.active, 0, 1) {
		r.lg.Debug("doStartR.enter")
		r.wgQuit.Add(2)
		GoExecute(r.loopFire, "fire")
		GoExecute(r.loopRead, "read")
		r.lg.Debug("doStartR.leave")
	}
}

func (r *tLineFileReader) Pause() {
	r.lg.Info("Pause")
	r.hasPause = true
}

func (r *tLineFileReader) Resume() {
	r.lg.Info("Resume")
	r.hasPause = false
}

func (r *tLineFileReader) doFire(data *string) {
	defer func() {
		if v := recover(); nil != v {
			fmt.Println(v)
		}
	}()

	r.listener(*data)
}

func (r *tLineFileReader) loopFire() {
	r.lg.Debug("loopFire.enter")

	defer func() {
		r.lg.Debug("loopFire.leave")
		r.wgQuit.Done()
	}()

	for {
		ref, ok := <-r.queue
		if ok {
			r.doFire(ref)
		} else {
			break
		}
	}
}

//if _, err := os.Stat(path); os.IsNotExist(err) {

func (r *tLineFileReader) open4Read() (ok bool) {
	ok = false

	r.lg.Debug("open4Read......")

	for r.IsActive() {
		inf, err := os.Stat(r.fpath)
		if nil == err && inf.Size() > 0 {
			break
		}

		r.lg.Debug("wait for file(%s) to prepare.......", r.fpath)
		time.Sleep(time.Millisecond * 500)
	}

	if !r.IsActive() {
		r.lg.Warn("read.quit for term")
		return
	}

	var err error
	if r.file, err = os.OpenFile(r.fpath, os.O_RDWR, 0666); nil != err {
		r.lg.Error("read.error %s", err.Error())
	} else {
		st, _ := os.Stat(r.fpath)
		r.fSize = int(st.Size())
		r.fd = int(r.file.Fd())
		if r.mMap() {
			for i := 0; i < len(rwKeyFlags); i++ {
				if r.mbb[i] != rwKeyFlags[i] {
					r.lg.Error("file format error %s", r.fpath)
					r.clean()
					break
				}
			}

			ok = nil != r.mbb
		}
	}

	return
}

func (r *tLineFileReader) file2Memory(rBuffer *TBufObj) {
	if err := syscall.Flock(r.fd, syscall.LOCK_EX); nil != err {
		return
	}

	defer syscall.Flock(r.fd, syscall.LOCK_UN)

	r.loadRWOffset()
	if r.offsetW > r.offsetR {
		for r.offsetR < r.offsetW && rBuffer.Len() < rwReadSize {
			rBuffer.WriteByte(r.mbb[rwDataPos16+r.offsetR])
			r.offsetR++
		}

		if r.IsActive() {
			r.saveRWOffset()
		}
	}
}

func (r *tLineFileReader) loopRead() {
	r.lg.Debug("loopRead.enter")

	defer func() {
		r.lg.Debug("loopRead.leave")

		if v := recover(); nil != v {
			r.lg.Error("error %v", v)
		}
		r.clean()
		r.wgQuit.Done()
	}()

	if ! r.open4Read() {
		r.lg.Error("open4Read.err")
		return
	}
	r.lg.Info("open4Read.OK")

	rBuffer := NewBufObj(rwReadSize)

	// [0 1 2 3, 4 5 6 7, 8 9 10 11, 12 13 14 15, 16 ]
	//  m o u s  r n v ?  head       tail       ***
	for r.IsActive() {
		if r.hasPause {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		if r.hasData() { // 第5个字节表示是否有数据
			r.file2Memory(rBuffer)

			for r.IsActive() {
				if line, ok := rBuffer.readLine(); ok {
					r.queue <- &line
				} else {
					break
				}
			}

		} else {
			time.Sleep(time.Millisecond * 100)
		}
	}

}

// 写部分
func NewLineFileW(fpath string, kbs int) (ILineFileWriter) {
	if kbs < 1 {
		kbs = 1
	} else if kbs > 1024 {
		kbs = 1024
	}

	w := &tLineFileWriter{
		tLineFileBase: &tLineFileBase{
			lg:     log.GetLogger("LWriter"),
			wgQuit: &sync.WaitGroup{},
			fpath:  fpath,
			fname:  getName(fpath),
			fSize:  kbs * 1024,
			active: 0,
			queue:  make(chan *string, 1024),
		},
	}
	w.onStart = w.doStartW
	return w
}

func (w *tLineFileWriter) doStartW() {
	if atomic.CompareAndSwapInt32(&w.active, 0, 1) {
		w.lg.Info("doStartW.enter")
		w.wgQuit.Add(1)
		GoExecute(w.loopSave, "save")
		w.lg.Info("doStartW.leave")
	}
}

func (w *tLineFileWriter) Write(data string) {
	if (w.IsActive()) {
		line := data + "\r\n"
		w.queue <- &line
	} else {
		w.lg.Info("skip %s", data)
	}
}

func (w *tLineFileWriter) openWrite() (bool) {
	var err error

	// 检查是否有目录
	dir := path.Dir(w.fpath)
	if !hasFile(dir) {
		os.Mkdir(dir, 077)
	}

	if hasFile(w.fpath) {
		w.lg.Info("open file %s", w.fpath)
		if w.file, err = os.OpenFile(w.fpath, os.O_RDWR, 0666); nil != err {
			w.lg.Error("error on open: %s", err.Error())
			return false
		}
	} else {
		w.lg.Info("create file %s", w.fpath)
		if w.file, err = os.Create(w.fpath); nil != err {
			w.lg.Error("error on create: %s", err.Error())
			return false
		}

		size := int64(w.fSize)
		w.file.Truncate(size)
	}

	w.lg.Info("use file %s", w.fpath)

	w.fd = int(w.file.Fd())

	if !w.mMap() {
		return false
	}

	for i := 0; i < len(rwKeyFlags); i++ {
		w.mbb[i] = rwKeyFlags[i]
	}
	return true
}

func (w *tLineFileWriter) memory2File(writeBuf *bytes.Buffer) {
	//w.lg.Debug("mem2file")
	// 这里开始写入数据
	// [0 1 2 3, 4 5 6 7, 8 9 10 11, 12 13 14 15, 16 ]
	//  m o u s  v ?      head       tail         ***
	// 锁文件，准备写, 这里的代码，对windows无效
	if err := syscall.Flock(w.fd, syscall.LOCK_EX); nil != err {
		return
	}

	defer syscall.Flock(w.fd, syscall.LOCK_UN)

	w.loadRWOffset()

	// 后面的空间不够，将数据挪到前面
	freeOnTail := uint32(w.fSize) - rwDataPos16 - w.offsetW
	dataSize := uint32(writeBuf.Len())
	if dataSize > freeOnTail {
		posW := 0
		for r := w.offsetR; r < w.offsetW; r++ {
			w.mbb[rwDataPos16+posW] = w.mbb[rwDataPos16+r]
			posW++
		}

		w.offsetW -= w.offsetR
		w.offsetR = 0
	}

	// 写入数据
	dstBuf := w.mbb[rwDataPos16+w.offsetW:]
	writeBuf.Read(dstBuf)
	writeBuf.Reset()

	w.offsetW += dataSize

	// 调整偏移量
	w.saveRWOffset()
}

func (w *tLineFileWriter) loopSave() {
	w.lg.Debug("loopSave.enter")

	tk := time.NewTicker(time.Millisecond * 10)

	defer func() {
		if v := recover(); nil != v {
			w.lg.Error("error on loopSave %v", v)
		}
		tk.Stop()
		w.clean()
		w.lg.Debug("loopSave.leave")
		w.wgQuit.Done()
	}()

	if !w.openWrite() {
		w.lg.Error("")
		return
	}

	wBuffer := bytes.NewBuffer(make([]byte, 81920))
	wBuffer.Reset()

	for {
		select {
		case ref, ok := <-w.queue:
			if !ok {
				if wBuffer.Len() > 0 {
					w.memory2File(wBuffer)
				}
				return
			}

			line := *ref
			wBuffer.WriteString(line)
			if wBuffer.Len() > 81920 {
				w.memory2File(wBuffer)
			}

		case <-tk.C:
			if wBuffer.Len() > 0 {
				w.memory2File(wBuffer)
			} else {
				// 没有数据，但是读写位置不一致
				if !w.hasData() && (w.offsetW != w.offsetR) {
					w.offsetR = w.offsetW
					w.saveRWOffset()
				}
			}
		}
	}
}

func TestRW() {
	log.Setup(log.NewDefConfig())
	fPath := "ggg.log"
	fSize := 1024 * 128
	lg := log.GetLogger("RW")

	var wSeq, rSeq int64

	w := NewLineFileW(fPath, fSize)

	genLine := func() {
		lg.Info("tipLine.enter")
		for i := 0; w.IsActive(); i++ {
			line := fmt.Sprintf("line-[%d]", i)
			w.Write(line)
			wSeq++
			//lg.Debug("send %s", line)
			time.Sleep(time.Millisecond * 5)
		}
	}

	r := NewLineFileReader(fPath, func(line string) {
		rSeq++
		//lg.Info("recv %s", line)
	})

	tipSync := func() {
		lg.Info("tipRead.enter")
		for r.IsActive() {
			lg.Info("W=%d desc=%s", wSeq, w.String())
			lg.Info("R=%d desc=%s\r\n", rSeq, r.String())
			time.Sleep(time.Second)
		}
	}

	tipDemo := func() {
		lg.Info("demo-========================")
	}

	go func() {
		for {
			GoExecute(tipDemo, "demo")
			time.Sleep(time.Second * 5)
		}
	}()

	w.Start()
	r.Start()
	GoExecute(genLine, "genLine")
	GoExecute(tipSync, "tipSync")

	exitSignals := make(chan os.Signal)
	signal.Notify(exitSignals,
		syscall.SIGINT,  // 2
		syscall.SIGTERM, // 15
		syscall.SIGQUIT, // 3
		syscall.SIGKILL, // 9
	)
	<-exitSignals
	close(exitSignals)

	w.Stop()
	r.Stop()
}
