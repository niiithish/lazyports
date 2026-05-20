package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/niiithish/lazyports/internal/ports"
	"github.com/niiithish/lazyports/internal/version"
)

func (m Model) View() string {
	if m.width == 0 {
		return appStyle.Render(" Loading lazyports… ")
	}

	if m.helpOpen {
		return appStyle.Width(m.width).Height(m.height).Align(lipgloss.Left, lipgloss.Top).Render(
			lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
				helpOverlayStyle.Render(helpText)),
		)
	}

	lay := computeLayout(m.width, m.height)
	body := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderPortsPanel(lay),
		m.renderDetailPanel(lay),
	)
	status := m.renderStatusBar(lay)
	if m.showUpdatePrompt() {
		status = m.renderUpdateBanner(lay) + "\n" + status
	}

	return strings.TrimRight(body, "\n") + "\n" + status
}

func (m Model) showUpdatePrompt() bool {
	return m.updateInfo.Available && !m.updateDismissed && m.mode == modeNormal && !m.helpOpen
}

func (m Model) renderUpdateBanner(lay layout) string {
	msg := fmt.Sprintf(
		" update v%s available (you have v%s) · U: show install command · esc: dismiss ",
		m.updateInfo.Latest,
		m.updateInfo.Current,
	)
	return statusBarStyle.Width(lay.width).Render(warnStyle.Render(msg))
}

func colPad(s string, width int) string {
	n := lipgloss.Width(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

func (m Model) portsPanelTitle(lay layout) string {
	mode := dimStyle.Render("dev")
	if m.showSystem {
		mode = warnStyle.Render("all tcp")
	}
	count := dimStyle.Render(fmt.Sprintf("%d", len(m.filtered)))
	return fmt.Sprintf("Ports · %s · %s", mode, count)
}

func (m Model) detailPanelTitle(p *ports.Port) string {
	if p == nil {
		return "Details"
	}
	return fmt.Sprintf("%s:%d", truncateHost(p.DisplayAddress()), p.Port)
}

func truncateHost(addr string) string {
	addr = strings.TrimPrefix(addr, "localhost")
	addr = strings.TrimPrefix(addr, "127.0.0.1")
	addr = strings.TrimPrefix(addr, "[::1]")
	addr = strings.TrimPrefix(addr, "::1")
	if addr == "" {
		return "localhost"
	}
	return strings.TrimPrefix(addr, ":")
}

func (m Model) renderPortsPanel(lay layout) string {
	const (
		colType = 2
		colPort = 7
		colPID  = 7
	)

	header := strings.Join([]string{
		colPad(tableHeaderStyle.Render("TP"), colType),
		colPad(tableHeaderStyle.Render("PORT"), colPort),
		colPad(tableHeaderStyle.Render("PID"), colPID),
		tableHeaderStyle.Render("PROCESS"),
	}, "")

	lines := []string{
		panelTitleStyle.Render(m.portsPanelTitle(lay)),
		header,
		dimStyle.Render(strings.Repeat("─", lay.leftInnerW)),
	}

	if len(m.filtered) == 0 {
		lines = append(lines, dimStyle.Render("  no matching ports"))
	} else {
		end := m.listOffset + lay.listH
		if end > len(m.filtered) {
			end = len(m.filtered)
		}
		procW := lay.leftInnerW - colType - colPort - colPID
		if procW < 8 {
			procW = 8
		}
		for i := m.listOffset; i < end; i++ {
			lines = append(lines, m.renderPortRow(m.filtered[i], i == m.cursor, colType, colPort, colPID, procW, lay.leftInnerW))
		}
	}

	lines = fitLines(lines, lay.bodyInnerH)

	return activePanelStyle.
		Width(lay.leftInnerW).
		Height(lay.bodyInnerH).
		Align(lipgloss.Left, lipgloss.Top).
		Render(strings.Join(lines, "\n"))
}

func (m Model) renderPortRow(p ports.Port, selected bool, colType, colPort, colPID, procW, innerW int) string {
	typeChar := "U"
	if p.Protocol == "TCP" {
		typeChar = "T"
	}
	proto := colPad(protoStyle(p.Protocol).Render(typeChar), colType)

	portNum := colPad(portStyle.Render(fmt.Sprintf("%d", p.Port)), colPort)

	pid := colPad(dimStyle.Render("—"), colPID)
	if p.PID > 0 {
		pid = colPad(fmt.Sprintf("%d", p.PID), colPID)
	}

	proc := processLabel(p)
	if proc == "" {
		proc = dimStyle.Render("unknown")
	} else {
		proc = processStyle.Render(truncate(proc, procW))
	}

	line := strings.Join([]string{proto, portNum, pid, proc}, "")

	if selected {
		return selectedRowStyle.Render(colPad(line, innerW))
	}
	return rowStyle.Render(line)
}

func processLabel(p ports.Port) string {
	if p.Process != "" {
		return p.Process
	}
	if p.PID > 0 {
		return fmt.Sprintf("pid:%d", p.PID)
	}
	return ""
}

func (m Model) renderDetailPanel(lay layout) string {
	p := m.selected()

	lines := []string{panelTitleStyle.Render(m.detailPanelTitle(p))}
	if p == nil {
		lines = append(lines, dimStyle.Render("select a port from the list"))
	} else {
		lines = append(lines,
			infoStyle.Render(fmt.Sprintf("%s %s", p.Protocol, p.State)),
			portStyle.Render(p.DisplayAddress()),
		)
		if p.PID > 0 {
			lines = append(lines, processStyle.Render(fmt.Sprintf("PID %d · %s", p.PID, p.Process)))
		}
		if p.User != "" {
			lines = append(lines, dimStyle.Render("User: ")+p.User)
		}
		if p.Uptime != "" {
			lines = append(lines, dimStyle.Render("Uptime: ")+p.Uptime)
		}
		if p.Command != "" {
			lines = append(lines, "", dimStyle.Render("Command:"))
			lines = append(lines, strings.Split(wrapText(p.Command, lay.rightInnerW), "\n")...)
		}
	}

	lines = fitLines(lines, lay.bodyInnerH)

	return inactivePanelStyle.
		Width(lay.rightInnerW).
		Height(lay.bodyInnerH).
		Align(lipgloss.Left, lipgloss.Top).
		Render(strings.Join(lines, "\n"))
}

func wrapText(s string, width int) string {
	if width < 10 {
		width = 10
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	var lines []string
	var cur strings.Builder
	for _, w := range words {
		if cur.Len() == 0 {
			cur.WriteString(w)
			continue
		}
		if cur.Len()+1+len(w) > width {
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(w)
		} else {
			cur.WriteString(" ")
			cur.WriteString(w)
		}
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderStatusBar(lay layout) string {
	var left string
	switch {
	case m.mode == modeConfirmRestart:
		left = warnStyle.Render(fmt.Sprintf(
			" Restart %s on %s? [y/N] ",
			m.confirmPort.Process, m.confirmPort.DisplayAddress(),
		))
	case m.mode == modeConfirmKill:
		left = warnStyle.Render(fmt.Sprintf(" Kill %s (PID %d)? [y/N] ", m.confirmPort.Process, m.confirmPort.PID))
	case m.mode == modeConfirmForceKill:
		left = errorStyle.Render(fmt.Sprintf(" Force kill %s (PID %d)? [y/N] ", m.confirmPort.Process, m.confirmPort.PID))
	case m.statusMsg != "" && time.Now().Before(m.statusExpiry):
		left = statusAccentStyle.Render(" " + m.statusMsg + " ")
	default:
		left = optionsStyle.Render(m.optionsText())
	}

	info := infoStyle.Render(fmt.Sprintf(" lazyports v%s ", version.Current()))
	if m.mode == modeFilter {
		m.filterInput.Width = min(lay.width-lipgloss.Width(" filter: ")-lipgloss.Width(info)-2, 40)
		filterLine := filterPromptStyle.Render("filter: ") + m.filterInput.View()
		return statusBarStyle.Width(lay.width).Render(filterLine + strings.Repeat(" ", max(1, lay.width-lipgloss.Width(filterLine)-lipgloss.Width(info))) + info)
	}

	if lipgloss.Width(left)+lipgloss.Width(info) <= lay.width {
		gap := lay.width - lipgloss.Width(left) - lipgloss.Width(info)
		if gap < 1 {
			gap = 1
		}
		return statusBarStyle.Width(lay.width).Render(left + strings.Repeat(" ", gap) + info)
	}

	return statusBarStyle.Width(lay.width).Render(left + "\n" + info)
}

func (m Model) optionsText() string {
	return " j/k: navigate, /: filter, a: system, s: restart, d: kill, r: refresh, ?: help, q: quit "
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
