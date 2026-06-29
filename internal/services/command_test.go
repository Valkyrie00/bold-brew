package services

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestExecuteCommand_CapturesStdout(t *testing.T) {
	var buf bytes.Buffer
	cmd := exec.Command("echo", "hello world")
	err := ExecuteCommand(cmd, &buf)
	if err != nil {
		t.Fatalf("ExecuteCommand() error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "hello world" {
		t.Errorf("output = %q, want %q", got, "hello world")
	}
}

func TestExecuteCommand_CapturesStderr(t *testing.T) {
	var buf bytes.Buffer

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}
	cmd = exec.Command("sh", "-c", "echo error >&2")

	err := ExecuteCommand(cmd, &buf)
	if err != nil {
		t.Fatalf("ExecuteCommand() error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "error" {
		t.Errorf("stderr output = %q, want %q", got, "error")
	}
}

func TestExecuteCommand_ReturnsError(t *testing.T) {
	var buf bytes.Buffer
	cmd := exec.Command("false")
	err := ExecuteCommand(cmd, &buf)
	if err == nil {
		t.Error("expected error from failed command")
	}
}

func TestExecuteCommand_InvalidCommand(t *testing.T) {
	var buf bytes.Buffer
	cmd := exec.Command("nonexistent-command-xyz-12345")
	err := ExecuteCommand(cmd, &buf)
	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

func TestExecuteCommand_MultiLineOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	var buf bytes.Buffer
	cmd := exec.Command("sh", "-c", "echo line1; echo line2; echo line3")
	err := ExecuteCommand(cmd, &buf)
	if err != nil {
		t.Fatalf("ExecuteCommand() error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("got %d lines, want 3: %q", len(lines), buf.String())
	}
}
