package ci

import (
	"bufio"
	"os/exec"
)

func DeactivateAllPorts() {
	deactivateCmd := exec.Command("port", "deactivate", "active")
	deactivateCmd.Start()
	deactivateCmd.Wait()
}

func ListSubports(port string) ([]string, error) {
	listCmd := exec.Command("port", "-q", "info", "--index", "--line", "--name", port, "subportof:"+port)
	stdout, err := listCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = listCmd.Start(); err != nil {
		return nil, err
	}
	subports := make([]string, 0, 1)
	stdoutScanner := bufio.NewScanner(stdout)
	for stdoutScanner.Scan() {
		line := stdoutScanner.Text()
		subports = append(subports, line)
	}
	if err = listCmd.Wait(); err != nil {
		return nil, err
	}
	return subports, nil
}
