//go:build linux

package ports

import (
	"os"
	"os/exec"
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

func TestResolveExecutableArgvKeepsRealBinary(t *testing.T) {
	t.Parallel()

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	argv := []string{exe, "-test.run=TestResolveExecutableArgvKeepsRealBinary"}
	got := resolveExecutableArgv(os.Getpid(), argv)
	if got[0] != exe {
		t.Fatalf("expected %q, got %q", exe, got[0])
	}
}

func TestResolveExecutableArgvFixesDisplayTitle(t *testing.T) {
	t.Parallel()

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	argv := []string{
		"puma 7.2.0 (tcp://localhost:3000) [otterkin-web]",
		"-C",
		"config/puma.rb",
	}
	got := resolveExecutableArgv(os.Getpid(), argv)
	if got[0] != exe {
		t.Fatalf("expected %q, got %q", exe, got[0])
	}
	if got[1] != "-C" || got[2] != "config/puma.rb" {
		t.Fatalf("unexpected argv tail: %v", got[1:])
	}
}

func TestLooksLikeExecutable(t *testing.T) {
	t.Parallel()

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	if !looksLikeExecutable(exe) {
		t.Fatalf("expected executable path %q to be recognized", exe)
	}
	if looksLikeExecutable("puma 7.2.0 (tcp://localhost:3000) [otterkin-web]") {
		t.Fatal("display title should not look executable")
	}
	if looksLikeExecutable("") {
		t.Fatal("empty path should not look executable")
	}

	if goExe, err := exec.LookPath("go"); err == nil && !looksLikeExecutable(goExe) {
		t.Fatalf("expected go binary %q to be recognized", goExe)
	}
}
