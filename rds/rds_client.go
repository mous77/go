package rds

import (
	"lib/log"
	"github.com/go-redis/redis"
	"sync/atomic"
	"time"
	"fmt"
)

type RdsClient struct {
	client *redis.Client
	lg     log.ILogger
	config *RdsConfig
	active int32
}

func NewRdsClient(_conf *RdsConfig) (*RdsClient) {
	return &RdsClient{
		lg:     log.GetLogger("redis."+_conf.String()),
		config: _conf,
	}
}

func (this *RdsClient) Start() {
	if atomic.CompareAndSwapInt32(&this.active, 0, 1) {
		//this.lg.Info("start")
		conf := this.config

		option := &redis.Options{
			Addr:     conf.Addr(),
			DB:       conf.DB,
			Password: conf.Auth,
		}

		this.client = redis.NewClient(option)
	}
}

func (this *RdsClient) Stop() {
	if atomic.CompareAndSwapInt32(&this.active, 1, 0) {
		//this.lg.Info("stop")
		this.client.Close()
	}
}

func (this *RdsClient) GetServer() (string) {
	return this.config.Addr();
}

func (this *RdsClient) IsActive() (bool) {
	return atomic.LoadInt32(&this.active) == 1
}

func (this *RdsClient)Keys(_pattern string)[]string{
	return this.client.Keys(_pattern).Val()
}

func (this *RdsClient) BLpop(_timeout time.Duration, _key string) ([]string) {
	res := this.client.BLPop(_timeout, _key)
	return res.Val()
}

func (this RdsClient) RPush(_key string, _val ... interface{}) {
	this.client.RPush(_key, _val...)
}

func (this *RdsClient) Publish(_chn string, _msg string) {
	this.client.Publish(_chn, _msg)
}

func (this *RdsClient) HMset(_key string, _val map[string]interface{}) {
	this.client.HMSet(_key, _val)
}

func (this *RdsClient) Set(_key string, _val interface{}, _secs int) {
	this.client.Set(_key, _val, time.Second*(time.Duration(_secs)))
}

func (this *RdsClient) Expire(_key string, _secs int) {
	this.client.Expire(_key, time.Duration(_secs)*time.Second)
}

func (this *RdsClient) Incr(_key string) {
	this.client.Incr(_key)
}

func (this *RdsClient) HIncrBy(_key string, _field string, _val int64) {
	this.client.HIncrBy(_key, _field, _val)
}

func (this RdsClient) Get(_key string) (string) {
	return this.client.Get(_key).Val()
}

func (this *RdsClient) hset(_key string, _field string, _val interface{}) {
	this.client.HSet(_key, _field, _val)
}

func (this *RdsClient) hget(_key string, _field string) (string) {
	return this.client.HGet(_key, _field).Val()
}

func (this *RdsClient) HMget(_key string, _fields ... string) ([]interface{}) {
	return this.client.HMGet(_key, _fields...).Val()
}

func (this *RdsClient) HGetAll(_key string) (map[string]string) {
	return this.client.HGetAll(_key).Val()
}

func (this *RdsClient) Del(_key string) {
	this.client.Del(_key)
}

func (this *RdsClient) HDel(_key string, _field string) {
	this.client.HDel(_key, _field)
}

func (this *RdsClient) Pipeline() (redis.Pipeliner) {
	return this.client.Pipeline()
}

func (this *RdsClient) Subscribe() (*redis.PubSub) {
	return this.client.Subscribe("CH_4_INIT_ONLY")
}

func TestRds() {
	this := NewRdsClient(NewRdsConf())
	this.Start()
	val := this.BLpop(time.Second, "key")
	for i, v := range val {
		var k int = i
		fmt.Printf("%d=%s\this\n", k, v)
	}

	//os.Exit(0)
}
