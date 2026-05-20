//go:build linux

package ports

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// enrichPorts fills missing PID/process info via /proc/net socket inodes.
func enrichPorts(ports []Port) []Port {
	inodes := buildInodeMap()
	if len(inodes) == 0 {
		return ports
	}

	out := make([]Port, len(ports))
	for i, p := range ports {
		if p.PID > 0 {
			out[i] = p
			continue
		}

		key := socketKey(p.Protocol, p.Address, p.Port)
		if inode, ok := inodes[key]; ok {
			if pid := findPIDByInode(inode); pid > 0 {
				p.PID = pid
				p.Process = processName(pid)
				p.User = processUser(pid)
				p.Command = processCommand(pid)
				p.Uptime = processUptime(pid)
			}
		}
		out[i] = p
	}
	return out
}

func buildInodeMap() map[string]string {
	m := make(map[string]string)
	for _, entry := range []struct {
		proto string
		file  string
	}{
		{"TCP", "tcp"},
		{"TCP", "tcp6"},
		{"UDP", "udp"},
		{"UDP", "udp6"},
	} {
		parseProcNet(filepath.Join("/proc/net", entry.file), entry.proto, m)
	}
	return m
}

func parseProcNet(path, proto string, out map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "sl") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		local := fields[1]
		inode := fields[9]
		addr, port, err := decodeProcAddr(local)
		if err != nil {
			continue
		}
		out[socketKey(proto, addr, port)] = inode
	}
}

func decodeProcAddr(hexAddr string) (string, int, error) {
	parts := strings.Split(hexAddr, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("bad addr")
	}

	port64, err := strconv.ParseUint(parts[1], 16, 16)
	if err != nil {
		return "", 0, err
	}
	port := int(port64)

	ipBytes, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", 0, err
	}

	var ip net.IP
	switch len(ipBytes) {
	case 4:
		ip = net.IP{ipBytes[3], ipBytes[2], ipBytes[1], ipBytes[0]}
	case 16:
		for i := 0; i < 16; i += 4 {
			binary.BigEndian.PutUint32(ipBytes[i:i+4], binary.LittleEndian.Uint32(ipBytes[i:i+4]))
		}
		ip = net.IP(ipBytes)
	default:
		return "", 0, fmt.Errorf("bad ip len")
	}

	return ip.String(), port, nil
}

func socketKey(proto, addr string, port int) string {
	if addr == "" {
		addr = "0.0.0.0"
	}
	return fmt.Sprintf("%s-%s-%d", proto, normalizeAddr(addr), port)
}

func normalizeAddr(addr string) string {
	if addr == "::" {
		return "0.0.0.0"
	}
	if ip := net.ParseIP(addr); ip != nil && ip.To4() == nil && ip.IsUnspecified() {
		return "0.0.0.0"
	}
	return addr
}

func findPIDByInode(inode string) int {
	target := "socket:[" + inode + "]"
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		fdDir := filepath.Join("/proc", e.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			if link == target {
				return pid
			}
		}
	}
	return 0
}

// FilterListenTCP returns ports that are TCP sockets in LISTEN state.
func FilterListenTCP(ports []Port) []Port {
	var out []Port
	for _, p := range ports {
		if p.Protocol == "TCP" && p.State == "LISTEN" {
			out = append(out, p)
		}
	}
	return out
}
