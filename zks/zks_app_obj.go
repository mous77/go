package zks

import(
	"strings"
)

// 应用对象
type TZksAppObj struct {
	Type   string
	Key    string
	Cfg    string
	Active bool
}

type TOnApp func(*TZksAppObj)

func NewZksAppObj(_type, _key, _cfg string) (*TZksAppObj) {
	if len(_type) + len(_key) + len(_cfg) == 0 {
		panic("no support nil args")
	}
	return &TZksAppObj{Type:_type, Key:_key, Cfg:_cfg, Active:true}
}

func (this *TZksAppObj)isSame(_other *TZksAppObj) (bool) {
	return (nil != _other) &&
			(0 == strings.Compare(_other.Type, this.Type) )&&
			(0 == strings.Compare(_other.Key, this.Key)) &&
			(0 == strings.Compare(_other.Cfg, this.Cfg))

}
