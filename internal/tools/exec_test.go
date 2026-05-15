package tools

import (
	"context"
	"testing"
	"time"
)

func TestRunCapturesStdout(t *testing.T) {
	got, err := Run(context.Background(), time.Second, "printf", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Stdout) != "hello" {
		t.Fatalf("stdout = %q", got.Stdout)
	}
}

func TestRunTimesOut(t *testing.T) {
	_, err := Run(context.Background(), time.Millisecond, "sleep", "1")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
