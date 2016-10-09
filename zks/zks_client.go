package zks

import (
	"fmt"
	"time"
	"sync/atomic"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/mous77/go/logger"
)

// 客户端封装对象
type TZksClient struct {
	log       logger.ILogger
	conn      *zk.Conn
	active    int32
	zkAddr    []string
	root      string
	curApp    *TZksAppObj
	needTypes map[string]*tZksAppType
	onApp     TOnApp
}

func NewZksClient(_zks []string, _cur *TZksAppObj, _root string, _on_app TOnApp, _types ...string) (*TZksClient) {
	types := make(map[string]*tZksAppType)

	client := &TZksClient{active:0, zkAddr:_zks, root:_root, curApp:_cur,
		needTypes:types, onApp:_on_app, log:logger.GetLogger("zksClient")}

	if len(_types) > 0 {
		for _, name := range _types {
			if len(name) > 0 {
				if _, ok := types[name]; !ok {
					types[name] = newAppType(client, name)
				}
			}
		}
	}

	return client
}

func makePath(conn *zk.Conn, path string) {
	if ok, _, _ := conn.Exists(path); !ok {
		conn.Create(path, nil, 0, nil)
	}
}

func (this *TZksClient)fireApp(_app *TZksAppObj) {
	if !_app.isSame(this.curApp) {
		this.onApp(_app)
	}
}

func (this *TZksClient)launch() {
	log := this.log

	var (
		conn *zk.Conn
		events <-chan zk.Event
		err error
	)
	if conn, events, err = zk.Connect(this.zkAddr, time.Second * 5); nil != err {
		panic(err)
	} else {
		conn.SetLogger(log)
		this.conn = conn

		log.Debug("evts", <-events)
		makePath(conn, this.root)

		app := this.curApp
		if nil != app {
			path := fmt.Sprintf("%s/%s", this.root, app.Type)
			makePath(conn, path)

			path = fmt.Sprintf("%s/%s/%s", this.root, app.Type, app.Key)
			conn.Create(path, []byte(app.Cfg), 1, nil)
		}

		for _, t := range this.needTypes {
			if cur_keys, stat, evts, e := conn.ChildrenW(t.path); nil != e {
				log.Error("error on childW, %s", e)
			} else {
				log.Debug("key ", cur_keys, stat, evts)
			}
		}
	}
}

func (this *TZksClient)Star() {
	if atomic.CompareAndSwapInt32(&this.active, 0, 1) {
		go this.launch()
	}
}

func (this *TZksClient)Stop() {
	if atomic.CompareAndSwapInt32(&this.active, 1, 0) {
		this.conn.Close()
	}
}