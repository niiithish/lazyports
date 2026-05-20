package ports

import "testing"

func TestDedupePortsSamePIDAndPort(t *testing.T) {
	t.Parallel()

	ports := []Port{
		{Protocol: "TCP", State: "LISTEN", Address: "127.0.0.1", Port: 3000, PID: 80719, Process: "ruby"},
		{Protocol: "TCP", State: "LISTEN", Address: "::1", Port: 3000, PID: 80719, Process: "ruby"},
	}

	out := DedupePorts(ports)
	if len(out) != 1 {
		t.Fatalf("expected 1 row, got %d: %+v", len(out), out)
	}
	if out[0].Address != "127.0.0.1" {
		t.Fatalf("expected localhost binding, got %s", out[0].Address)
	}
}

func TestDedupePortsDifferentPIDs(t *testing.T) {
	t.Parallel()

	ports := []Port{
		{Protocol: "TCP", State: "LISTEN", Address: "127.0.0.1", Port: 3000, PID: 1, Process: "a"},
		{Protocol: "TCP", State: "LISTEN", Address: "127.0.0.1", Port: 3000, PID: 2, Process: "b"},
	}

	out := DedupePorts(ports)
	if len(out) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(out))
	}
}
