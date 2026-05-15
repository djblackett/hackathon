package tools

import "os/exec"

func Available(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
