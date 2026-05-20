//go:build linux

package ports

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// RestartPortProcess stops the process on a port and starts it again with the same command.
func RestartPortProcess(p Port) (int, error) {
	if p.PID <= 0 {
		return 0, fmt.Errorf("no process to restart")
	}

	argv, err := processArgv(p.PID)
	if err != nil {
		return 0, fmt.Errorf("read command: %w", err)
	}

	cwd, err := processCWD(p.PID)
	if err != nil {
		return 0, fmt.Errorf("read working directory: %w", err)
	}

	env := processEnv(p.PID)
	oldPID := p.PID
	argv = resolveExecutableArgv(oldPID, argv)

	if err := stopProcess(oldPID, 3*time.Second); err != nil {
		return 0, err
	}

	newPID, err := startDetached(argv, cwd, env)
	if err != nil {
		return 0, fmt.Errorf("start process: %w", err)
	}

	return newPID, nil
}

func processArgv(pid int) ([]string, error) {
	data, err := os.ReadFile(procFile(pid, "cmdline"))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty cmdline")
	}

	var argv []string
	for _, part := range bytes.Split(data, []byte{0}) {
		if len(part) == 0 {
			continue
		}
		argv = append(argv, string(part))
	}
	if len(argv) == 0 {
		return nil, fmt.Errorf("empty cmdline")
	}
	return argv, nil
}

func processExe(pid int) (string, error) {
	exe, err := os.Readlink(procFile(pid, "exe"))
	if err != nil {
		return "", err
	}
	exe = strings.TrimSpace(exe)
	if exe == "" {
		return "", fmt.Errorf("empty exe")
	}
	if idx := strings.Index(exe, " (deleted)"); idx >= 0 {
		exe = exe[:idx]
	}
	return exe, nil
}

// resolveExecutableArgv fixes argv[0] when a runtime rewrites it for display.
// Ruby/Puma, Node, and others often set argv[0] to something like
// "puma 7.2.0 (tcp://localhost:3000) [app]" which is not a real binary path.
func resolveExecutableArgv(pid int, argv []string) []string {
	if len(argv) == 0 {
		return argv
	}

	if looksLikeExecutable(argv[0]) {
		return argv
	}

	exe, err := processExe(pid)
	if err != nil {
		return argv
	}

	out := make([]string, len(argv))
	copy(out, argv)
	out[0] = exe
	return out
}

func looksLikeExecutable(path string) bool {
	if path == "" || strings.Contains(path, " ") {
		return false
	}

	if strings.Contains(path, "/") {
		info, err := os.Stat(path)
		if err != nil {
			return false
		}
		if info.IsDir() {
			return false
		}
		return info.Mode()&0111 != 0
	}

	found, err := exec.LookPath(path)
	return err == nil && found != ""
}

func processCWD(pid int) (string, error) {
	cwd, err := os.Readlink(procFile(pid, "cwd"))
	if err != nil {
		return "", err
	}
	if cwd == "" {
		return "", fmt.Errorf("empty cwd")
	}
	return cwd, nil
}

func processEnv(pid int) []string {
	data, err := os.ReadFile(procFile(pid, "environ"))
	if err != nil || len(data) == 0 {
		return os.Environ()
	}

	var env []string
	for _, part := range bytes.Split(data, []byte{0}) {
		if len(part) == 0 {
			continue
		}
		env = append(env, string(part))
	}
	if len(env) == 0 {
		return os.Environ()
	}
	return env
}

func stopProcess(pid int, timeout time.Duration) error {
	if err := Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal PID %d: %w", pid, err)
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !IsAlive(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := Kill(pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("force kill PID %d: %w", pid, err)
	}

	time.Sleep(200 * time.Millisecond)
	if IsAlive(pid) {
		return fmt.Errorf("process %d still running", pid)
	}
	return nil
}

func startDetached(argv []string, cwd string, env []string) (int, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = cwd
	cmd.Env = env

	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return 0, err
	}
	defer devNull.Close()

	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return 0, err
	}

	go func() {
		_ = cmd.Wait()
	}()

	return cmd.Process.Pid, nil
}

// RestartCommand returns the command that would be re-run for a process.
func RestartCommand(pid int) string {
	argv, err := processArgv(pid)
	if err != nil {
		return processCommand(pid)
	}
	argv = resolveExecutableArgv(pid, argv)
	return strings.Join(argv, " ")
}

// ExecutablePath returns the resolved binary path for a PID.
func ExecutablePath(pid int) (string, error) {
	exe, err := processExe(pid)
	if err != nil {
		return "", err
	}
	return filepath.Clean(exe), nil
}
