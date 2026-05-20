package ports

import (
	"os/user"
	"strings"
)

// Well-known system ports to hide from the dev-focused view.
var systemPorts = map[int]struct{}{
	21: {}, 22: {}, 25: {}, 53: {}, 67: {}, 68: {}, 69: {},
	80: {}, 110: {}, 123: {}, 143: {}, 443: {}, 445: {}, 465: {},
	587: {}, 631: {}, 993: {}, 995: {},
	5353: {}, 5355: {}, 5357: {}, 5900: {},
}

var systemProcessHints = []string{
	"systemd",
	"dnsmasq",
	"tailscaled",
	"sshd",
	"cupsd",
	"avahi",
	"networkmanager",
	"systemd-resolved",
	"resolved",
	"rpcbind",
	"gssproxy",
	"sssd",
	"colord",
	"modemmanager",
	"udisks",
	"upower",
	"polkit",
	"dbus-daemon",
	"pipewire",
	"wireplumber",
	"snapd",
	"containerd",
	"dockerd",
	"fail2ban",
	"ntpd",
	"chronyd",
}

// devProcessHints are processes we always treat as development servers.
var devProcessHints = []string{
	"node",
	"npm",
	"pnpm",
	"yarn",
	"bun",
	"deno",
	"python",
	"ruby",
	"rails",
	"puma",
	"java",
	"go",
	"air",
	"cargo",
	"rust",
	"php",
	"dotnet",
	"vite",
	"webpack",
	"next",
	"nuxt",
	"astro",
	"remix",
	"svelte",
	"docker-proxy",
	"rootlesskit",
	"kubectl",
	"port-forward",
	"cloudflared",
	"ngrok",
	"serve",
	"http.server",
	"uvicorn",
	"gunicorn",
	"flask",
	"django",
	"fastapi",
	"redis-server",
	"mysqld",
	"postgres",
	"mongod",
	"sqlite",
	"esbuild",
	"turbo",
}

// ScanDev returns TCP listen ports owned by dev processes (your user), excluding system services.
func ScanDev() ([]Port, error) {
	all, err := ScanAll()
	if err != nil {
		return nil, err
	}
	return FilterDevPorts(all), nil
}

// FilterDevPorts keeps user-owned development servers and drops system noise.
func FilterDevPorts(ports []Port) []Port {
	me, _ := user.Current()
	var out []Port
	for _, p := range ports {
		if IsDevPort(p, me) {
			out = append(out, p)
		}
	}
	return DedupePorts(out)
}

// IsDevPort reports whether a port looks like a local development server.
func IsDevPort(p Port, me *user.User) bool {
	if p.Protocol != "TCP" || p.State != "LISTEN" {
		return false
	}
	if p.PID <= 0 {
		return false
	}

	if _, blocked := systemPorts[p.Port]; blocked {
		return false
	}

	identity := strings.ToLower(strings.Join([]string{p.Process, p.Command}, " "))
	if isSystemProcess(identity) {
		return false
	}

	// Docker port forwards and similar run as root but are dev-relevant.
	if isDevProcess(identity) {
		return true
	}

	// Default: only show ports owned by the current user.
	if me != nil && p.User != "" {
		if p.User != me.Username {
			return false
		}
	}

	// User-owned listener on a high port (typical dev range).
	if p.Port >= 1024 && p.User != "" && p.User != "root" {
		return true
	}

	return false
}

func isSystemProcess(identity string) bool {
	for _, hint := range systemProcessHints {
		if strings.Contains(identity, hint) {
			return true
		}
	}
	return false
}

func isDevProcess(identity string) bool {
	for _, hint := range devProcessHints {
		if strings.Contains(identity, hint) {
			return true
		}
	}
	return false
}
