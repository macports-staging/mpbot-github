package ci

import (
	"os/exec"
	"path"
)

type buildWorker struct {
	worker
	portChan chan string
}

func newBuildWorker(session *Session) *buildWorker {
	return &buildWorker{
		worker:   worker{session: session, quitChan: make(chan byte)},
		portChan: make(chan string),
	}
}

func (worker *buildWorker) start() {
	go func() {
		returnCode := byte(0)
		for {
			select {
			case port := <-worker.portChan:
				if port == "" {
					worker.quitChan <- returnCode
					return
				}
				subports, err := ListSubports(port)
				if err != nil {
					returnCode = 1
					GlobalLogger.LogTextChan <- &LogText{"port-" + port + "-subports-fail", []byte(err.Error())}
					continue
				}
				for _, subport := range subports {
					DeactivateAllPorts()
					portTmpDir := path.Join(worker.session.tmpDir, subport)
					logFilename := path.Join(worker.session.tmpDir, "port-"+subport+"-dep-install.log")
					err := mpbbToLog("install-dependencies", subport, portTmpDir, logFilename)
					statusString := "success"
					if err != nil {
						if eerr, ok := err.(*exec.ExitError); ok {
							if !eerr.Success() {
								returnCode = 1
								statusString = "fail"
							}
						}
					}
					GlobalLogger.LogFileChan <- &LogFile{
						"port-" + subport + "-dep-summary-" + statusString,
						path.Join(portTmpDir, "logs/dependencies-progress.txt"),
					}
					// TODO: upload to Minio?
					GlobalLogger.LogBigFileChan <- &LogFile{
						"port-" + port + "-dep-install-output-" + statusString,
						logFilename,
					}
					if err != nil {
						continue
					}

					logFilename = path.Join(worker.session.tmpDir, "port-"+subport+"-install.log")
					mpbbToLog("install-port", subport, portTmpDir, logFilename)
					statusString = "success"
					if err != nil {
						if eerr, ok := err.(*exec.ExitError); ok {
							if !eerr.Success() {
								returnCode = 1
								statusString = "fail"
							}
						}
					}
					GlobalLogger.LogFileChan <- &LogFile{
						"port-" + subport + "-install-summary-" + statusString,
						path.Join(portTmpDir, "logs/ports-progress.txt"),
					}
					GlobalLogger.LogBigFileChan <- &LogFile{
						"port-" + port + "-install-output-" + statusString,
						logFilename,
					}
				}
			}
		}
	}()
}
