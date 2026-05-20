//go:build linux

package ports

import (
	"os"
	"strings"
	"testing"
)

func TestProcessArgvCurrentProcess(t *testing.T) {
	t.Parallel()

	argv, err := processArgv(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if len(argv) == 0 {
		t.Fatal("expected argv")
	}
	if !strings.Contains(strings.Join(argv, " "), "test") {
		t.Fatalf("unexpected argv: %v", argv)
	}
}

func TestProcessCWDCurrentProcess(t *testing.T) {
	t.Parallel()

	cwd, err := processCWD(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if cwd == "" {
		t.Fatal("expected cwd")
	}
}

func TestRestartCommandCurrentProcess(t *testing.T) {
	t.Parallel()

	cmd := RestartCommand(os.Getpid())
	if cmd == "" {
		t.Fatal("expected command")
	}
}
