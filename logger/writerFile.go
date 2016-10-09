package logger

import (
	"fmt"
	"os"
	"bufio"
	"time"
)

type tFileWriter struct {
	file   *os.File
	writer *bufio.Writer
	ymd    string
	lines  int
}

func (this *tFileWriter)Recv(_line string) {
	if !manager.config.GenFile {
		return
	}

	if 0 == len(this.ymd) {
		this.Open()
	}

	if 0 < len(this.ymd) {
		this.writer.WriteString(_line)
		this.lines++
	}
}

const (
	dateFormat = "2006-01-02"
	//timeFormat = "(2006-01-02 15.04.05)"
)

func (this *tFileWriter)Flush() {
	if this.lines > 0 {
		this.writer.Flush()
		this.lines = 0
	}

	ymd := time.Now().Format(dateFormat)
	if ymd != this.ymd {
		if len(this.ymd) > 0 {
			fmt.Printf("close log of today: %s \r\n", ymd)
			this.Close()
		}
	}
}

func fileExists(_file_name string) (bool) {
	exists := true
	if _, err := os.Stat(_file_name); os.IsNotExist(err) {
		exists = false
	}
	return exists
}

func (this *tFileWriter)Open() {
	now := time.Now()
	ymd := now.Format(dateFormat)
	file_dir := fmt.Sprintf("%s/logs", manager.config.Root)

	if !fileExists(file_dir) {
		fmt.Printf("new dir %s\r\n", file_dir)
		os.MkdirAll(file_dir, 0777)
	}

	file_name := fmt.Sprintf("%s/%s %s.log", file_dir, manager.config.App, ymd)

	var err error
	if fileExists(file_name) {
		if this.file, err = os.OpenFile(file_name, os.O_APPEND, 0666); nil == err {
			fmt.Printf("open log file %s ok!r\\n", file_name)
		}
	} else {
		if this.file, err = os.Create(file_name); nil == err {
			fmt.Printf("Create log file %s ok!\r\n", file_name)
		}
	}

	if nil != err {
		fmt.Printf("err: %s\r\n", err.Error())
	} else {
		this.writer = bufio.NewWriter(this.file)
		this.ymd = ymd

		this.file.WriteString("== open ==\r\n")
	}
}

func (this *tFileWriter)Close() {
	if nil != this.writer {
		fmt.Printf("close %s\r\n", this.file.Name())
		this.writer.Flush()
		this.file.Close()

		this.ymd = ""
		this.file = nil
		this.writer = nil
	}
}
