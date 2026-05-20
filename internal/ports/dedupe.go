package ports

import (
	"fmt"
)

// DedupePorts collapses multiple socket bindings for the same process into one row.
// Puma/Rails often bind both 127.0.0.1:3000 and [::1]:3000 — we show one entry.
func DedupePorts(ports []Port) []Port {
	if len(ports) == 0 {
		return ports
	}

	groups := make(map[string][]Port)
	var order []string

	for _, p := range ports {
		key := dedupeKey(p)
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], p)
	}

	out := make([]Port, 0, len(order))
	for _, key := range order {
		out = append(out, pickCanonical(groups[key]))
	}

	sortPorts(out)
	return out
}

func dedupeKey(p Port) string {
	if p.PID > 0 {
		return fmt.Sprintf("%s-%d-%d", p.Protocol, p.Port, p.PID)
	}
	return fmt.Sprintf("%s-%d-%s", p.Protocol, p.Port, p.Address)
}

func pickCanonical(group []Port) Port {
	if len(group) == 1 {
		return group[0]
	}

	best := group[0]
	bestScore := addressScore(best.Address)
	for _, p := range group[1:] {
		if score := addressScore(p.Address); score > bestScore {
			best = p
			bestScore = score
		}
	}
	return best
}

func addressScore(addr string) int {
	switch addr {
	case "127.0.0.1", "::1":
		return 100
	case "0.0.0.0", "::":
		return 80
	default:
		if len(addr) >= 4 && addr[:4] == "127." {
			return 90
		}
		return 10
	}
}
