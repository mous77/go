package zks

import(
	"fmt"
)

// 需要关注的应用类型
type tZksAppType struct {
	owner *TZksClient
	name  string
	path  string
	apps  map[string]*TZksAppObj
}

func newAppType(_owner *TZksClient, _name string) (*tZksAppType) {
	path := fmt.Sprintf("%s/%s", _owner.root, _name)
	maps := make(map[string]*TZksAppObj)
	return &tZksAppType{owner:_owner, name:_name, path:path, apps:maps}
}

func (this *tZksAppType)updateApps(_cur_keys []string) {

	// 先认为全部应用挂掉
	for _, app := range this.apps {
		app.Active = false
	}

	// 再次激活找到的应用
	for _, key := range _cur_keys {
		if app, ok := this.apps[key]; ok {
			app.Active = true
		} else {
			path := fmt.Sprintf("%s/%s", this.path, key)
			if data, _, err := this.owner.conn.Get(path); nil != err {
				this.owner.log.Error("error on get %s", err.Error())
			} else {
				app = NewZksAppObj(this.name, key, string(data))
				this.apps[key] = app
				this.owner.fireApp(app)
			}
		}
	}

	// 然后清除未活动应用
	for key, app := range this.apps {
		if !app.Active {
			this.owner.onApp(app)
			delete(this.apps, key)
		}
	}

}

