package ci

type worker struct {
	session  *Session
	quitChan chan byte
}

func (worker *worker) wait() byte {
	return <-worker.quitChan
}
