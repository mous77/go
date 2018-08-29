package main

import (
	"github.com/mous77/go/logger"
	"time"
)

type tShower struct {
	cont [20]string
	idx  int
}

func (s *tShower) Recv(_msg string) {
	s.cont[s.idx] = _msg
	s.idx = (s.idx + 1) % 20
}

func main() {
	s := &tShower{}

	go func() {
		for ; ; {
			for _, s := range s.cont {
				println(s)
			}

			println("==============")
			time.Sleep(time.Second)
		}
	}()

	logger.TestLog(s)

}
