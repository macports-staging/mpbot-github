package ci

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"strings"
)

// Workers are not exported.
type Session struct {
	tmpDir       string
	ports        []string
	buildCounter uint32
	// TODO: split to lintResults and buildResults or use unified struct type
	// TODO: Return error in Run() if build failed
	// Collects lint and build failure
	results chan string
}

func NewSession() (*Session, error) {
	tmpDir, err := ioutil.TempDir("", "ci-build-")
	if err != nil {
		return nil, err
	}
	ports, err := GetPortList()
	if err != nil {
		// No cleanup in CI
		//os.RemoveAll(tmpDir)
		return nil, err
	}
	return &Session{
		tmpDir:  tmpDir,
		ports:   ports,
		results: make(chan string),
	}, nil
}

func (session *Session) Run() error {
	GlobalLogger.LogTextChan <- &LogText{"port-list", []byte(strings.Join(session.ports, "\n"))}
	var err error
	bWorker := newBuildWorker(session)
	bWorker.start()
	go func() {
		for _, port := range session.ports {
			bWorker.portChan <- port
		}
		bWorker.portChan <- ""
	}()

	{
		lintArgs := make([]string, len(session.ports)+2)
		lintArgs[0] = "-p"
		lintArgs[1] = "lint"
		copy(lintArgs[2:], session.ports)
		lintCmd := exec.Command("port", lintArgs...)
		out, err := lintCmd.CombinedOutput()
		statusString := "success"
		if err != nil {
			if eerr, ok := err.(*exec.ExitError); ok {
				if !eerr.Success() {
					statusString = "fail"
					err = errors.New("lint failed")
				}
			}
		}
		GlobalLogger.LogTextChan <- &LogText{"port-lint-output-" + statusString, out}
	}

	if bWorker.wait() != 0 && err == nil {
		err = errors.New("build failed")
	}
	return err
}
