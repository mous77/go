package log

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

func (fw *tFileWriter)Recv(_msg string) {
	if !manager.config.GenFile {
		return
	}

	if 0 == len(fw.ymd) {
		fw.Open()
	}

	if 0 < len(fw.ymd) {
		fw.writer.WriteString(_msg)
		fw.writer.WriteString("\r\n")
		fw.lines++
	}
}

const (
	dateFormat = "2006-01-02"
	//timeFormat = "(2006-01-02 15.04.05)"
)

func (fw *tFileWriter)Flush() {
	if fw.lines > 0 {
		fw.writer.Flush()
		fw.lines = 0
	}

	ymd := time.Now().Format(dateFormat)
	if ymd != fw.ymd {
		if len(fw.ymd) > 0 {
			fmt.Printf("close log of today: %s \r\n", ymd)
			fw.Close()
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

func (fw *tFileWriter)Open() {
	now := time.Now()
	ymd := now.Format(dateFormat)
	file_dir := fmt.Sprintf("%s/logs", manager.config.Root)

	if !fileExists(file_dir) {
		fmt.Printf("new dir %s\r\n", file_dir)
		os.MkdirAll(file_dir, 0777)
	}

	file_name := fmt.Sprintf("%s/%s.%s.log", file_dir, manager.config.App, ymd)

	var err error
	if fileExists(file_name) {
		if fw.file, err = os.OpenFile(file_name, os.O_APPEND, 0666); nil == err {
			fmt.Printf("open log file %s ok!\r\n", file_name)
		}
	} else {
		if fw.file, err = os.Create(file_name); nil == err {
			fmt.Printf("Create log file %s ok!\r\n", file_name)
		}
	}

	if nil != err {
		fmt.Printf("err: %s\r\n", err.Error())
	} else {
		fw.writer = bufio.NewWriter(fw.file)
		fw.ymd = ymd

		fw.file.WriteString("== open ==\r\n")
	}
}

func (fw *tFileWriter)Close() {
	if nil != fw.writer {
		fmt.Printf("close %s\r\n", fw.file.Name())
		fw.writer.Flush()
		fw.file.Close()

		fw.ymd = ""
		fw.file = nil
		fw.writer = nil
	}
}
