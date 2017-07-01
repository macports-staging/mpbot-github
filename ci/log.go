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
	LogBigFileChan chan *LogFile
	LogFileChan    chan *LogFile
	LogTextChan    chan *LogText
	mimeWriter     *multipart.Writer
	quitChan       chan byte
	stopChan       chan byte
	remoteLog      *remoteLogger
}

type LogFile struct {
	FieldName, Filename string
}

type LogText struct {
	FieldName string
	Text      []byte
}

func newLogger(w io.Writer) *Logger {
	logger := &Logger{
		mimeWriter:     multipart.NewWriter(w),
		quitChan:       make(chan byte),
		stopChan:       make(chan byte),
		LogBigFileChan: make(chan *LogFile, 4),
		LogFileChan:    make(chan *LogFile, 4),
		LogTextChan:    make(chan *LogText, 4),
	}
	logger.remoteLog = newRemoteLogger(logger)
	return logger
}

func (l Logger) Run() {
	go func() {
		for {
			select {
			case <-l.stopChan:
				l.mimeWriter.Close()
				l.quitChan <- 0
				return
			case logFile := <-l.LogBigFileChan:
				l.remoteLog.logBigFileChan <- logFile
			case logFile := <-l.LogFileChan:
				if logFile == nil {
					l.remoteLog.logBigFileChan <- nil
					continue
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
					l.remoteLog.logBigFileChan <- nil
					continue
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

func (l Logger) Quit() byte {
	l.remoteLog.logBigFileChan <- nil
	l.remoteLog.wait()
	return <-l.quitChan
}
