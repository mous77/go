package logger

type tLogWriters struct {
	manager *tLogManager
	items   map[string]ILogWriter
}

func newWriters() *tLogWriters {
	return &tLogWriters{items:make(map[string]ILogWriter)}
}

func (this *tLogWriters)setup() {
	this.items["file"] = &tFileWriter{file:nil, writer:nil}
	this.items["cons"] = &tConsoleWriter{enable:manager.config.Console}
}

func (this *tLogWriters)close() {
	for _, w := range this.items {
		w.Close()
	}
}

func (this *tLogWriters)recv(_item *tLogItem) {
	line := _item.String()
	for _, w := range this.items {
		w.Recv(line)
	}
}

func (this *tLogWriters)flush() {
	for _, w := range this.items {
		w.Flush()
	}
}
