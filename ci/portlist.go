package ci

import (
	"bufio"
	"os/exec"
	"regexp"
)

func GetPortList() ([]string, error) {
	gitCmd := exec.Command("git", "diff", "--name-status", "macports/master...HEAD", "--")
	stdout, err := gitCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = gitCmd.Start(); err != nil {
		return nil, err
	}
	ports := make([]string, 0, 1)
	gitRegexp := regexp.MustCompile(`[AM]\t[^\._/][^/]*/([^/]+)/Portfile`)
	stdoutScanner := bufio.NewScanner(stdout)
	for stdoutScanner.Scan() {
		line := stdoutScanner.Text()
		if match := gitRegexp.FindStringSubmatch(line); match != nil {
			ports = append(ports, match[1])
		}
	}
	if err = gitCmd.Wait(); err != nil {
		return nil, err
	}
	return ports, nil
}
