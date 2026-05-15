package tools

import (
	"bytes"
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
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return CommandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, fmt.Errorf("%s timed out after %s", name, timeout)
	}
	if err != nil {
		return CommandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
	}
	return CommandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, nil
}
