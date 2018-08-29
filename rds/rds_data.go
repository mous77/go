package rds

import (
	"fmt"
)

type RdsData struct {
	Topic string
	Body  string
}

func NewRdsData(_topic string, _body string) (*RdsData) {
	return &RdsData{Topic: _topic, Body: _body}
}

var NilRdsData = &RdsData{}

func (d *RdsData) String() (string) {
	return fmt.Sprintf("%s:%s", d.Topic, d.Body)
}

type OnRdsData func(*RdsData)
