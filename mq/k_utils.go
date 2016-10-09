package mq

import (
	"github.com/samuel/go-zookeeper/zk"
	"time"
	"fmt"
	"encoding/json"
	"strconv"
	"errors"
	"github.com/mous77/go/logger"
)

const (
	c_KFS_BROKER_IDS = "/brokers/ids"
)

/*
  给定zookeeper地址列表 获取 broker 地址列表
 */
func GetKBrokers(_zks []string) (brokers []string, err error) {
	lg := logger.GetLogger("k_utils")
	conn, _, e1 := zk.Connect(_zks, time.Second * 3)
	if e1 != nil {
		err = errors.New(fmt.Sprintf("can not connect zks:%s, err:%s", _zks, e1.Error()))
	}else {
		defer func(){
			lg.Info("close zks connection %s", _zks)
			conn.Close()
		}()

		conn.SetLogger(lg)

		ids, _, e2 := conn.Children(c_KFS_BROKER_IDS)
		if nil != e2 {
			err = errors.New(fmt.Sprintf("error on get %s, err:%s", c_KFS_BROKER_IDS, e2.Error()))
		} else {
			brokers = make([]string, len(ids))
			for i, k := range ids {
				path := fmt.Sprintf(c_KFS_BROKER_IDS + "/%s", k)
				data, _, e3 := conn.Get(path)
				if nil!= e3{
					err = errors.New(fmt.Sprintf("error on get %s, err:%s", path, e3.Error()))
				}else{
					var obj map[string]interface{}
					json.Unmarshal(data, &obj)
					var host string = obj["host"].(string)
					var port float64 = obj["port"].(float64)
					broker := fmt.Sprintf("%s:%s", host,  strconv.FormatFloat(port, 'f', 0, 64) );
					brokers[i] = broker
				}
			}
		}
	}
	return
}

func TestZKS(_lg logger.ILogger){
	zks := []string{"127.0.0.1:2181","127.0.0.1:2182"}
	brks,err := GetKBrokers(zks)
	if nil!=err {
		_lg.Error("error on getBrokers %s, \r\n\t%s", zks, err.Error())
	}else{
		_lg.Info("brokers=%s",brks)
	}
}
