package ci

import (
	"io"
	"mime/multipart"
	"os"
	"time"
)

var GlobalLogger *Logger = newLogger(os.Stdout)

func init() {
	GlobalLogger.Run()
}

type Logger struct {
	LogFileChan chan *LogFile
	LogTextChan chan *LogText
	mimeWriter  *multipart.Writer
	quitChan    chan byte
}

type LogFile struct {
	FieldName, Filename string
}

type LogText struct {
	FieldName string
	Text      []byte
}

func newLogger(w io.Writer) *Logger {
	return &Logger{
		mimeWriter:  multipart.NewWriter(w),
		quitChan:    make(chan byte),
		LogFileChan: make(chan *LogFile, 4),
		LogTextChan: make(chan *LogText, 4),
	}
}

func (l Logger) Run() {
	go func() {
		for {
			select {
			case logFile := <-l.LogFileChan:
				if logFile == nil {
					l.mimeWriter.Close()
					l.quitChan <- 0
					return
				}
				writer, err := l.mimeWriter.CreateFormField(logFile.FieldName)
				if err != nil {
					continue
				}
				file, err := os.Open(logFile.Filename)
				if err != nil {
					continue
				}

				io.Copy(writer, file)

				file.Close()
				os.Remove(logFile.Filename)
			case logText := <-l.LogTextChan:
				if logText == nil {
					l.mimeWriter.Close()
					l.quitChan <- 0
					return
				}
				writer, err := l.mimeWriter.CreateFormField(logText.FieldName)
				if err != nil {
					continue
				}
				_, err = writer.Write(logText.Text)
			case <-time.After(time.Minute * 5):
				l.LogTextChan <- &LogText{"keep-alive", []byte("I'm working hard!")}
			}
		}
	}()
}

func (l Logger) Wait() byte {
	return <-l.quitChan
}
