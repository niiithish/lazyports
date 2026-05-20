package ports

import (
	"testing"
)

func TestParseAddressPort(t *testing.T) {
	tests := []struct {
		in      string
		addr    string
		port    int
		wantErr bool
	}{
		{"0.0.0.0:3000", "0.0.0.0", 3000, false},
		{"127.0.0.1:8080", "127.0.0.1", 8080, false},
		{"[::]:3000", "::", 3000, false},
		{"[::1]:5432", "::1", 5432, false},
	}

	for _, tt := range tests {
		addr, port, err := parseAddressPort(tt.in)
		if tt.wantErr && err == nil {
			t.Fatalf("%s: expected error", tt.in)
		}
		if !tt.wantErr && err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.in, err)
		}
		if addr != tt.addr || port != tt.port {
			t.Fatalf("%s: got %s:%d want %s:%d", tt.in, addr, port, tt.addr, tt.port)
		}
	}
}

func TestScanIntegration(t *testing.T) {
	ports, err := ScanAll()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(ports) == 0 {
		t.Fatal("expected at least one port")
	}
	for _, p := range ports {
		if p.Port == 3000 {
			t.Logf("port 3000: %+v", p)
		}
	}
}

func TestSSLineRegex(t *testing.T) {
	line := `tcp   LISTEN 0      5                                0.0.0.0:3000       0.0.0.0:*    users:(("python3",pid=12902,fd=10))`
	m := ssLineRe.FindStringSubmatch(line)
	if m == nil {
		t.Fatal("regex did not match")
	}
	if m[1] != "tcp" || m[2] != "LISTEN" || m[3] != "0.0.0.0:3000" {
		t.Fatalf("unexpected groups: %v", m)
	}
	pm := ssProcRe.FindStringSubmatch(m[4])
	if pm == nil || pm[1] != "python3" || pm[2] != "12902" {
		t.Fatalf("unexpected process groups: %v", pm)
	}
}
