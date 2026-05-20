package ports

import (
	"os/user"
	"testing"
)

func TestFilterDevPortsExcludesSystemPortNumbers(t *testing.T) {
	t.Parallel()
	me := &user.User{Username: "dev"}
	ports := []Port{
		{Protocol: "TCP", State: "LISTEN", Port: 53, PID: 100, Process: "foo", User: "dev"},
		{Protocol: "TCP", State: "LISTEN", Port: 3000, PID: 101, Process: "node", User: "dev", Command: "node server.js"},
	}
	out := FilterDevPorts(ports)
	if len(out) != 1 || out[0].Port != 3000 {
		t.Fatalf("expected only port 3000, got %+v", out)
	}
	if !IsDevPort(out[0], me) {
		t.Fatal("expected 3000 to be dev port")
	}
}

func TestFilterDevPortsExcludesSystemProcesses(t *testing.T) {
	t.Parallel()
	me := &user.User{Username: "dev"}
	p := Port{Protocol: "TCP", State: "LISTEN", Port: 9000, PID: 1, Process: "systemd-resolved", User: "systemd-resolve"}
	if IsDevPort(p, me) {
		t.Fatal("systemd-resolved should not be dev port")
	}
}

func TestFilterDevPortsIncludesDockerProxy(t *testing.T) {
	t.Parallel()
	me := &user.User{Username: "dev"}
	p := Port{Protocol: "TCP", State: "LISTEN", Port: 3000, PID: 1, Process: "docker-proxy", User: "root", Command: "docker-proxy -proto tcp"}
	if !IsDevPort(p, me) {
		t.Fatal("docker-proxy on 3000 should be dev port")
	}
}

func TestFilterDevPortsRequiresPID(t *testing.T) {
	t.Parallel()
	me := &user.User{Username: "dev"}
	p := Port{Protocol: "TCP", State: "LISTEN", Port: 3000, PID: 0}
	if IsDevPort(p, me) {
		t.Fatal("port without pid should be excluded")
	}
}
