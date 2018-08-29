package log

import "fmt"

// console log writer
type tConsoleWriter struct {
	enable bool
}

func (cw *tConsoleWriter) Open() {
}

func (cw *tConsoleWriter) Recv(_msg string) {
	if cw.enable {
		fmt.Println(_msg)
	}
}

func (cw *tConsoleWriter) Flush() {
}

func (cw *tConsoleWriter) Close() {
}

// custom log writer
type tCustomWriter struct {
	handler ILogHandler
}

func (uw *tCustomWriter) Open() {

}

func (uw *tCustomWriter) Recv(_msg string) {
	uw.handler.Recv(_msg)
}

func (uw *tCustomWriter) Flush() {

}

func (uw *tCustomWriter) Close() {

}
