package logger

import "fmt"

type tConsoleWriter struct {
	enable bool
}

func (this *tConsoleWriter)Recv(_msg string) {
	if this.enable {
		fmt.Print(_msg)
	}
}

func (this *tConsoleWriter)Flush(){
}

func (this *tConsoleWriter)Close(){
}

func (this *tConsoleWriter)Open(){
}
