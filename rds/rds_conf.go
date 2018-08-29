package rds

import (
	"strings"
	"strconv"
	"fmt"
	"os"
)

type RdsConfig struct {
	Host string
	Port int

	DB   int
	Auth string
	Pool int
}

func NewRdsConf() (*RdsConfig) {
	return &RdsConfig{Host: "127.0.0.1", Port: 6379, Auth: "foobared", Pool: 1}
}

func (this *RdsConfig) Addr() (string) {
	return fmt.Sprintf("%s:%d", this.Host, this.Port)
}

func (this *RdsConfig) ByLine(_line string)(*RdsConfig) {
	ary := strings.Split(_line, ":")
	switch len(ary) {
	case 3:
		this.Host = ary[0]
		this.Port, _ = strconv.Atoi(ary[1])
		this.Auth = ary[2]
	case 4:
		this.Host = ary[0]
		this.Port, _ = strconv.Atoi(ary[1])
		this.Auth = ary[2]
		this.Pool, _ = strconv.Atoi(ary[3])
	default:
		fmt.Printf("invalid redis args %s \r\n", _line)
		os.Exit(0)
	}
	return this
}

func (this *RdsConfig) ByMap(_map map[string]interface{}) {
	if v, ok := _map["host"]; ok {
		this.Host = v.(string)
	}

	if v, ok := _map["port"]; ok {
		this.Port = v.(int)
	}

	if v, ok := _map["auth"]; ok {
		this.Auth = v.(string)
	}

	if v, ok := _map["db"]; ok {
		this.DB = v.(int)
	}

	if v, ok := _map["pool"]; ok {
		this.Pool = v.(int)
	}
}

func (this *RdsConfig) String() (string) {
	return fmt.Sprintf("%s:%d", this.Host, this.Port)
}
