package tools

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type CommandResult struct {
	Stdout []byte
	Stderr []byte
}

func Run(ctx context.Context, timeout time.Duration, name string, args ...string) (CommandResult, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	stdout, err := cmd.Output()
	stderr := []byte(nil)
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = exitErr.Stderr
	}
	if ctx.Err() == context.DeadlineExceeded {
		return CommandResult{Stdout: stdout, Stderr: stderr}, fmt.Errorf("%s timed out after %s", name, timeout)
	}
	if err != nil {
		return CommandResult{Stdout: stdout, Stderr: stderr}, err
	}
	return CommandResult{Stdout: stdout, Stderr: stderr}, nil
}
