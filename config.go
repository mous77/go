package app

import (
	"sync/atomic"
	"gopkg.in/yaml.v2"
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"lib/log"
	"errors"
)

type ANode struct {
	data map[interface{}]interface{}
}

func (n *ANode) getItem(_path string) (res interface{}) {
	res = nil
	root := n.data
	for {
		idx := strings.IndexByte(_path, '/')
		if idx < 0 {
			res, _ = root[_path]
			break
		}

		key := _path[:idx]
		val, _ := root[key]
		if nil == val {
			break
		}

		_path = _path[idx+1:]
		idx = strings.IndexByte(_path, '/')
		if idx < 0 {
			tmp := val.(map[interface{}]interface{})
			res, _ = tmp[_path]
			break
		}

		root, _ = val.(map[interface{}]interface{})
	}
	return
}

func (n *ANode) GetObj(_path string) (*ANode) {
	item := n.getItem(_path)
	return &ANode{item.(map[interface{}]interface{})}
}

func (this *ANode) GetStr(_path string, _def string) (string) {
	obj := this.getItem(_path)
	if str, ok := obj.(string); ok {
		return str
	} else {
		return _def
	}
}

func (this *ANode) GetInt(_path string, _def int) (int) {
	obj := this.getItem(_path)
	if i, ok := obj.(int); ok {
		return i
	} else {
		return _def
	}
}

func (this *ANode) GetBool(_path string, _def bool) (bool) {
	obj := this.getItem(_path)
	if b, ok := obj.(bool); ok {
		return b
	} else {
		return _def
	}
}

func (this *ANode) GetAry(_path string) (res []int) {
	obj := this.getItem(_path)
	ary := obj.([]interface{})
	for _, val := range ary {
		res = append(res, val.(int))
	}
	return
}

func (this *ANode) GetList(_path string) ([]interface{}) {
	obj := this.getItem(_path)
	return obj.([]interface{})
}

type AConfig struct {
	loaded int32
	root   *ANode
	app    string
}

func NewAConfig(_name string) (*AConfig) {
	node := &ANode{make(map[interface{}]interface{})}
	return &AConfig{ 0, node, _name}
}

func (this *AConfig) LoadDef(_do_load func()) {
	dir, _ := os.Getwd()
	dir = strings.Replace(dir, "\\", "/", -1)
	def_yml := fmt.Sprintf("%s/%s.yaml", dir, this.app)
	this.LoadYaml(def_yml, _do_load)
}

func (this *AConfig) LoadYaml(_path string, _do_load func()) {
	if atomic.CompareAndSwapInt32(&this.loaded, 0, 1) {
		if buf, err := ioutil.ReadFile(_path); nil != err {
			e := errors.New(fmt.Sprintf("error on load yaml(%s) %s", _path, err.Error()))
			panic(e)
		} else {
			fmt.Printf("loadYaml(%s)\r\n", _path)
			yaml.Unmarshal(buf, this.root.data)
			this.loadLog()
			_do_load()
		}
	}
}

func (this *AConfig) loadLog() {
	lv_name := this.GetStr("log/level", "INFO")
	lgc := &log.TConfig{
		App:     this.app,
		Console: this.GetBool("log/console", true),
		Source:  this.GetBool("log/source", false),
		GenFile: this.GetBool("log/genFile", true),
		Level:   log.LevelByName(lv_name),
	}
	log.Setup(lgc)
}

func (this *AConfig) GetObj(_path string) (*ANode) {
	return this.root.GetObj(_path)
}

func (this *AConfig) GetStr(_path string, _def string) (string) {
	return this.root.GetStr(_path, _def)
}

func (this *AConfig) GetInt(_path string, _def int) (int) {
	return this.root.GetInt(_path, _def)
}

func (this *AConfig) GetBool(_path string, _def bool) (bool) {
	return this.root.GetBool(_path, _def)
}

func (this *AConfig) GetAry(_path string) ([]int) {
	return this.root.GetAry(_path)
}

func (this *AConfig) GetList(_path string) (res []*ANode) {
	lst := this.root.GetList(_path)
	for _, tmp := range lst {
		node := &ANode{tmp.(map[interface{}]interface{})}
		res = append(res, node)
	}
	return
}

type MyCfg struct {
	AConfig
	rdsHost string
	rdsPort int
	rdsPass string
}

func (m *MyCfg) doLoad() {
	m.rdsHost = m.GetStr("redis/host", "127.0.0.1")
	m.rdsPort = m.GetInt("redis/port", 6379)
	m.rdsPass = m.GetStr("redis/host", "foobared")
}

func TestConf() {
	fn := "/Users/mous/gopath/prj/ecpm/ccad/bin/ccax.yaml"

	ac := NewAConfig("caax")
	this := &MyCfg{AConfig: *ac}

	this.LoadYaml(fn, this.doLoad)

	x := this.GetObj("a/b")

	lst := this.GetList("ccau")

	fmt.Printf("this=%v %v %v\r\n", *this, x, lst)

	for _, u := range lst {
		fmt.Printf("%s:%d\r\n", u.GetStr("host", ""), u.GetInt("port", 0))
	}

	os.Exit(0)

}
