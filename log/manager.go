package log

import (
	"time"
	"sync"
)

type tLogManager struct {
	locks   *sync.Mutex
	insts   map[string]ILogger
	items   chan *tLogItem
	config  *TConfig
	active  int32
	writers map[string]ILogWriter
}

func newManager() *tLogManager {
	return &tLogManager{
		active:  0,
		locks:   &sync.Mutex{},
		insts:   make(map[string]ILogger),
		config:  NewDefConfig(),
		items:   make(chan *tLogItem, 1024),
		writers: make(map[string]ILogWriter),
	}
}

func (m *tLogManager) setup(_handler ILogHandler) {
	m.writers["cons"] = &tConsoleWriter{enable: m.config.Console}

	m.writers["file"] = &tFileWriter{file: nil, writer: nil}

	if nil != _handler {
		m.writers["user"] = &tCustomWriter{handler: _handler}
	}
}

func (this *tLogManager) dispatch(_item *tLogItem) {
	var msg string

	if nil != _item {
		msg = _item.String()
	}

	for _, w := range this.writers {
		if nil != _item {
			w.Recv(msg)
		} else {
			w.Flush()
		}
	}
}

func (m *tLogManager) run() {
	tk := time.NewTicker(time.Second)

	defer func() {
		for _, w := range m.writers {
			w.Close()
		}

		for k := range m.insts {
			delete(m.insts, k)
		}

		tk.Stop()
	}()

	terminated := false
	needFlush := false
	lastFlush := time.Now()

	for !terminated {
		select {
		case item := <-m.items:
			needFlush = true

			switch item.iType {
			case itDATA:
				m.dispatch(item)
			case itDONE:
				terminated = true
			}

		case nowTime := <-tk.C:
			if needFlush && lastFlush.Add(time.Second).After(nowTime) {
				m.dispatch(nil)
				lastFlush = nowTime
				needFlush = false
			}
		}
	}
}

func (m *tLogManager) getLogger(_key string) ILogger {
	defer m.locks.Unlock()
	m.locks.Lock()

	lg := m.insts[_key]
	if nil == lg {
		cfg := NewDefConfig()
		cfg.copyFrom(m.config)

		lg = &tLogInstance{
			config: cfg,
			key:    _key,
			items:  m.items,
		}

		m.insts[_key] = lg
	}
	return lg
}
