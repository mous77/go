package utils

import (
	"os/exec"
	"os"
	"path"
	"path/filepath"
	"strings"
	"bufio"
	"io"
	"fmt"
	"strconv"
)

type IIniLoader interface {
	Load(_ini_path string)
	GetIntVal(_sec string, _key string) (val int, ok bool)
	GetIntAry(_sec string, _key string) (val []int, ok bool)
	GetStrVal(_sec string, _key string) (val string, ok bool)
	GetSection(_sec string) (map[string]string)
}

type tIniLoader struct {
	iniPath  string
	sections map[string]*tIniSection
}

func NewIniLoader() (*tIniLoader) {
	return &tIniLoader{iniPath:getDefConfig(), sections:make(map[string]*tIniSection)}
}

func getCurDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func getDefConfig() (string) {
	bin_name, _ := exec.LookPath(os.Args[0])
	full_path, _ := filepath.Abs(bin_name)
	sufix := path.Ext(full_path)
	no_sufix := strings.Trim(full_path, sufix)
	return no_sufix + ".ini"
}

func (this *tIniLoader)Load(_path string) {
	if len(_path) > 0 {
		this.iniPath = _path

		if !IsFileExists(_path) && !strings.Contains(_path, "/"){
			this.iniPath =  getCurDir() +"/"+ _path
		}

	}
	lines := this.loadLines()
	this.parseLines(lines)
}

func (this *tIniLoader)String() (string) {
	ary := make([]string, 1024)[0:0]
	ary = append(ary, "file:" + this.iniPath)
	for _, sec := range this.sections {
		ary = append(ary, "[" + sec.name + "]")
		for key, ite := range sec.items {
			ary = append(ary, key + " = " + ite.value)
		}
	}

	lines := make([]byte, 8192)[0:]
	for _, line := range ary {
		buf := []byte(line + "\r\n")
		lines = append(lines, buf...)
	}
	return string(lines)
}

func (this *tIniLoader)getItem(_sec string, _key string) (ite *tIniItem, ok bool) {
	ok = false
	if sec, y := this.sections[_sec]; y {
		ite = sec.items[_key]
		ok = nil != ite
	}
	return
}

func (this *tIniLoader)GetIntVal(_sec string, _key string, _def int) (val int) {
	if ite, y := this.getItem(_sec, _key); y {
		val = ite.IntVal()
	} else {
		val = _def
	}
	return
}
func (this *tIniLoader)GetIntAry(_sec string, _key string, _def int) (val []int) {
	ok := false
	if ite, y := this.getItem(_sec, _key); y {
		val, ok = ite.IntAry()
	}

	if !ok {
		val = []int{_def}
	}

	return
}
func (this *tIniLoader)GetStrVal(_sec string, _key string, _def string) (val string) {
	if ite, y := this.getItem(_sec, _key); y {
		val = ite.StrVal()
	} else {
		val = _def
	}
	return
}

func (this *tIniLoader)GetSection(_sec string) (map[string]string) {
	kv := make(map[string]string)
	if sec, ok := this.sections[_sec]; ok {
		for k, v := range sec.items {
			kv[k] = v.value
		}
	}
	return kv
}

type tLine struct {
	index int
	data  string
}

func (this *tLine)err(_msg string) (string) {
	return fmt.Sprintf("error %s @line[%d] %s", _msg, this.index, this.data)
}

func (this *tLine)isSection() (name string, ok bool) {
	ok = false
	if strings.HasPrefix(this.data, "[") {
		if !strings.HasSuffix(this.data, "]") {
			println(this.err("unknow key"))
			return
		}
		buf := []byte(this.data)
		blen := len(buf)
		if blen < 2 {
			println(this.err("section error"))
			return
		}

		name = string(buf[1:blen - 1])
		ok = true
	}
	return
}

func (this *tLine)isItem() (key string, val string, ok bool) {
	ok = false
	idx := strings.IndexByte(this.data, '=')
	if idx < 1 {
		println(this.err("error "))
		return
	}

	buf := []byte(this.data)

	cdx := strings.Index(this.data, "//")
	if cdx < idx {
		cdx = len(buf)
	}

	key = string(buf[0:idx])
	val = strings.Trim(string(buf[idx + 1:cdx]), " ")
	ok = true

	return
}

func (this *tIniLoader)loadLines() ([]*tLine) {
	f, err := os.Open(this.iniPath)
	if nil != err {
		panic(err)
	}
	defer f.Close()

	lines := make([]*tLine, 1024)
	slice := lines[0:0]
	seq := 0

	addLine := func(line string){
		if len(line)==0{
			return
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 ||
				strings.HasPrefix(line, "#") ||
				strings.HasPrefix(line, "//") {
			return
		}
		slice = append(slice, &tLine{index:seq, data:line})
	}

	rd := bufio.NewReader(f)
	for {
		seq += 1
		line, err := rd.ReadString('\n')
		if nil != err || io.EOF == err {
			addLine(line)
			break
		}
		addLine(line)
	}

	return slice
}

func (this *tIniLoader)parseLines(_lines []*tLine) {
	var sec *tIniSection = nil

	for _, line := range _lines {
		if key, ok := line.isSection(); ok {
			sec = newSection(key)
			this.sections[key] = sec
		} else {
			if nil == sec {
				println(line.err("section not found"))
				continue
			} else {
				if key, val, ok := line.isItem(); ok {
					sec.addItem(key, val)
				}
			}
		}
	}
}

type tIniSection struct {
	name  string
	items map[string]*tIniItem
}

func newSection(_name string) (*tIniSection) {
	return &tIniSection{name:_name, items:make(map[string]*tIniItem)}
}

func (this *tIniSection)addItem(_key string, _val string) {
	this.items[_key] = newItem(_val)
}

type tIniItem struct {
	value string
}

func (this *tIniItem)IntVal() (int) {
	i, _ := strconv.Atoi(this.value)
	return i
}

func (this *tIniItem)IntAry() (ary []int, ok bool) {
	tmp := strings.Split(string(this.value), ",")
	for _, val := range tmp {
		if i, err := strconv.Atoi(val); nil == err {
			ary = append(ary, i)
			ok = true
		} else {
			println("error %s", err.Error())
		}
	}
	return
}

func (this *tIniItem)StrVal() (string) {
	return this.value
}

func newItem(_value string) (*tIniItem) {
	return &tIniItem{value:_value}
}

func IsFileExists(_file_name string) (bool) {
	exists := true
	if _, err := os.Stat(_file_name); os.IsNotExist(err) {
		exists = false
	}
	return exists
}