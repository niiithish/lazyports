package ports

import (
	"fmt"
	"os/user"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Port represents a local socket binding and the process that owns it.
type Port struct {
	Protocol  string
	State     string
	Address   string
	Port      int
	PID       int
	Process   string
	User      string
	Command   string
	FD        int
	Uptime    string
}

func (p Port) DisplayAddress() string {
	switch {
	case p.Address == "0.0.0.0" || p.Address == "*":
		return fmt.Sprintf("*:%d", p.Port)
	case p.Address == "127.0.0.1" || p.Address == "::1":
		return fmt.Sprintf("localhost:%d", p.Port)
	default:
		if strings.HasPrefix(p.Address, "[") {
			return fmt.Sprintf("%s:%d", p.Address, p.Port)
		}
		return fmt.Sprintf("%s:%d", p.Address, p.Port)
	}
}

func (p Port) Key() string {
	return fmt.Sprintf("%s-%s-%d-%d", p.Protocol, p.Address, p.Port, p.PID)
}

var (
	ssLineRe  = regexp.MustCompile(`^(\S+)\s+(\S+)\s+\d+\s+\d+\s+(\S+)\s+\S+(?:\s+(users:\(.+\)))?$`)
	ssProcRe  = regexp.MustCompile(`users:\(\("([^"]*)",pid=(\d+)(?:,fd=(\d+))?\)\)`)
)

// Scan returns development TCP listen ports (default view).
func Scan() ([]Port, error) {
	return ScanDev()
}

// ScanAll returns every socket from ss.
func ScanAll() ([]Port, error) {
	out, err := runSS()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var ports []Port

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Netid") {
			continue
		}

		m := ssLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		addrPort := m[3]
		address, portNum, err := parseAddressPort(addrPort)
		if err != nil {
			continue
		}

		pid := 0
		fd := 0
		process := ""
		if len(m) > 4 && m[4] != "" {
			pm := ssProcRe.FindStringSubmatch(m[4])
			if pm != nil {
				process = pm[1]
				pid, _ = strconv.Atoi(pm[2])
				if len(pm) > 3 && pm[3] != "" {
					fd, _ = strconv.Atoi(pm[3])
				}
			}
		}
		if process == "" && pid > 0 {
			process = processName(pid)
		}

		p := Port{
			Protocol: strings.ToUpper(m[1]),
			State:    m[2],
			Address:  address,
			Port:     portNum,
			PID:      pid,
			Process:  process,
			User:     processUser(pid),
			Command:  processCommand(pid),
			FD:       fd,
			Uptime:   processUptime(pid),
		}

		key := fmt.Sprintf("%s-%s-%d-%s", p.Protocol, p.Address, p.Port, p.State)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		ports = append(ports, p)
	}

	ports = enrichPorts(ports)
	sortPorts(ports)
	return ports, nil
}

// ScanListenTCP returns TCP ports in LISTEN state (the common dev use-case).
func ScanListenTCP() ([]Port, error) {
	all, err := ScanAll()
	if err != nil {
		return nil, err
	}
	return DedupePorts(FilterListenTCP(all)), nil
}

func sortPorts(ports []Port) {
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].Port != ports[j].Port {
			return ports[i].Port < ports[j].Port
		}
		if ports[i].Protocol != ports[j].Protocol {
			return ports[i].Protocol < ports[j].Protocol
		}
		return ports[i].Address < ports[j].Address
	})
}

func parseAddressPort(addrPort string) (string, int, error) {
	if strings.HasPrefix(addrPort, "[") {
		idx := strings.LastIndex(addrPort, "]:")
		if idx == -1 {
			return "", 0, fmt.Errorf("invalid address: %s", addrPort)
		}
		address := addrPort[1:idx]
		portStr := addrPort[idx+2:]
		port, err := strconv.Atoi(portStr)
		return address, port, err
	}

	parts := strings.Split(addrPort, ":")
	if len(parts) < 2 {
		return "", 0, fmt.Errorf("invalid address: %s", addrPort)
	}
	port, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return "", 0, err
	}
	address := strings.Join(parts[:len(parts)-1], ":")
	return address, port, nil
}

func processUser(pid int) string {
	if pid <= 0 {
		return ""
	}
	status, err := readProcFile(pid, "status")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(status, "\n") {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				uid, err := strconv.Atoi(fields[1])
				if err != nil {
					return ""
				}
				u, err := user.LookupId(strconv.Itoa(uid))
				if err != nil {
					return fields[1]
				}
				return u.Username
			}
		}
	}
	return ""
}
