//go:build linux

package ports

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func runSS() (string, error) {
	cmd := exec.Command("ss", "-tulnp")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if len(stderr.Bytes()) > 0 {
			return "", err
		}
	}
	return string(out), nil
}

func processName(pid int) string {
	cmdline, err := readProcFile(pid, "comm")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cmdline)
}

func processCommand(pid int) string {
	data, err := os.ReadFile(procFile(pid, "cmdline"))
	if err != nil {
		return processName(pid)
	}
	if len(data) == 0 {
		return processName(pid)
	}
	parts := strings.Split(string(bytes.ReplaceAll(data, []byte{0}, []byte(" "))), " ")
	return strings.TrimSpace(strings.Join(parts, " "))
}

func processUptime(pid int) string {
	if pid <= 0 {
		return ""
	}
	out, err := exec.Command("ps", "-o", "etime=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func procFile(pid int, file string) string {
	return fmt.Sprintf("/proc/%d/%s", pid, file)
}

func readProcFile(pid int, file string) (string, error) {
	data, err := os.ReadFile(procFile(pid, file))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Kill sends a signal to the process owning the port.
func Kill(pid int, signal syscall.Signal) error {
	if pid <= 0 {
		return fmt.Errorf("no process to kill")
	}
	return syscall.Kill(pid, signal)
}

// KillPortProcess terminates the process on a port with SIGTERM.
func KillPortProcess(p Port) error {
	return Kill(p.PID, syscall.SIGTERM)
}

// ForceKillPortProcess force-kills the process with SIGKILL.
func ForceKillPortProcess(p Port) error {
	return Kill(p.PID, syscall.SIGKILL)
}

// IsAlive checks whether a PID still exists.
func IsAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

// PIDs returns unique PIDs from a list of ports.
func PIDs(ports []Port) []int {
	seen := make(map[int]struct{})
	var pids []int
	for _, p := range ports {
		if p.PID <= 0 {
			continue
		}
		if _, ok := seen[p.PID]; ok {
			continue
		}
		seen[p.PID] = struct{}{}
		pids = append(pids, p.PID)
	}
	return pids
}

// FilterByPID returns ports owned by a PID.
func FilterByPID(ports []Port, pid int) []Port {
	var out []Port
	for _, p := range ports {
		if p.PID == pid {
			out = append(out, p)
		}
	}
	return out
}

// FormatPID is a helper for logging.
func FormatPID(pid int) string {
	return strconv.Itoa(pid)
}
