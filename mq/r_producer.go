package mq

type IRmqProducer interface {
	IMQWorker
	Setup(_key, _name_svr string)
	Push(_topic *string, _flag,  _key, _val []byte)
}
