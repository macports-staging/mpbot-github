package ci

import (
	"bufio"
	"io"
	"net"
	"os"
)

type remoteLogger struct {
	logBigFileChan chan *LogFile
	parent         *Logger
	quitChan       chan byte
}

func newRemoteLogger(parent *Logger) *remoteLogger {
	return &remoteLogger{
		logBigFileChan: make(chan *LogFile, 4),
		parent:         parent,
		quitChan:       make(chan byte),
	}
}

func (r remoteLogger) run() {
	go func() {
		for {
			select {
			case logFile := <-r.logBigFileChan:
				if logFile == nil {
					r.parent.stopChan <- 0
					r.quitChan <- 0
					return
				}
				file, err := os.Open(logFile.Filename)
				if err != nil {
					continue
				}
				conn, err := net.Dial("tcp", "termbin.com:9999")
				if err != nil {
					continue
				}
				io.Copy(conn, file)
				file.Close()
				os.Remove(logFile.Filename)
				url, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					continue
				}
				r.parent.LogTextChan <- &LogText{"upload-" + logFile.FieldName, []byte(url)}
			}
		}
	}()
}

func (r remoteLogger) wait() {
	<-r.quitChan
	return
}
