package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/niiithish/lazyports/internal/ports"
	"github.com/niiithish/lazyports/internal/update"
)

type mode int

const (
	modeNormal mode = iota
	modeFilter
	modeConfirmKill
	modeConfirmForceKill
	modeConfirmRestart
)

type tickMsg time.Time

type portsLoadedMsg struct {
	ports []ports.Port
	err   error
}

type killResultMsg struct {
	port  ports.Port
	force bool
	err   error
}

type restartResultMsg struct {
	port   ports.Port
	newPID int
	err    error
}

type updateCheckedMsg update.Info

// Model is the root bubbletea model.
type Model struct {
	allPorts         []ports.Port
	filtered         []ports.Port
	cursor           int
	filterInput      textinput.Model
	mode             mode
	statusMsg        string
	statusExpiry     time.Time
	width            int
	height           int
	refreshEvery     time.Duration
	lastRefresh      time.Time
	confirmPort      ports.Port
	helpOpen         bool
	listOffset       int
	showSystem       bool
	updateInfo       update.Info
	updateDismissed  bool
}

func New(refreshEvery time.Duration) Model {
	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 80
	ti.Prompt = ""
	ti.PromptStyle = filterPromptStyle
	ti.TextStyle = filterInputStyle
	ti.PlaceholderStyle = dimStyle
	ti.Cursor.Style = filterInputStyle

	return Model{
		filterInput:  ti,
		refreshEvery: refreshEvery,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(loadPorts(m.showSystem), tickCmd(m.refreshEvery), checkUpdateCmd())
}

func tickCmd(d time.Duration) tea.Cmd {
	if d <= 0 {
		return nil
	}
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadPorts(showSystem bool) tea.Cmd {
	return func() tea.Msg {
		var p []ports.Port
		var err error
		if showSystem {
			p, err = ports.ScanListenTCP()
		} else {
			p, err = ports.ScanDev()
		}
		return portsLoadedMsg{ports: p, err: err}
	}
}

func killCmd(p ports.Port, force bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if force {
			err = ports.ForceKillPortProcess(p)
		} else {
			err = ports.KillPortProcess(p)
		}
		return killResultMsg{port: p, force: force, err: err}
	}
}

func restartCmd(p ports.Port) tea.Cmd {
	return func() tea.Msg {
		newPID, err := ports.RestartPortProcess(p)
		return restartResultMsg{port: p, newPID: newPID, err: err}
	}
}

func checkUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		return updateCheckedMsg(update.Check(ctx))
	}
}

func (m *Model) resize() {
	lay := computeLayout(m.width, m.height)
	m.filterInput.Width = min(lay.leftInnerW-4, 48)
}

func (m *Model) moveCursor(idx int) {
	if len(m.filtered) == 0 {
		m.cursor = 0
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(m.filtered) {
		idx = len(m.filtered) - 1
	}
	m.cursor = idx
	m.syncListOffset()
}

func (m *Model) syncListOffset() {
	lay := computeLayout(m.width, m.height)
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	}
	if m.cursor >= m.listOffset+lay.listH {
		m.listOffset = m.cursor - lay.listH + 1
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		return m, nil

	case tea.KeyMsg:
		if m.helpOpen {
			m.helpOpen = false
			return m, nil
		}
		if m.mode == modeFilter {
			return m.updateFilter(msg)
		}
		if m.mode == modeConfirmKill || m.mode == modeConfirmForceKill || m.mode == modeConfirmRestart {
			return m.updateConfirm(msg)
		}
		return m.updateNormal(msg)

	case tickMsg:
		var cmds []tea.Cmd
		if m.refreshEvery > 0 && time.Since(m.lastRefresh) >= m.refreshEvery {
			m.lastRefresh = time.Now()
			cmds = append(cmds, loadPorts(m.showSystem))
		}
		cmds = append(cmds, tickCmd(m.refreshEvery))
		return m, tea.Batch(cmds...)

	case portsLoadedMsg:
		if msg.err != nil {
			m.setStatus("scan failed: "+msg.err.Error(), 5*time.Second)
			return m, nil
		}
		prevPID, prevPort, prevProto := 0, 0, ""
		if p := m.selected(); p != nil {
			prevPID, prevPort, prevProto = p.PID, p.Port, p.Protocol
		}
		m.allPorts = msg.ports
		m.applyFilter()
		m.lastRefresh = time.Now()

		if prevPID > 0 {
			for i, p := range m.filtered {
				if p.PID == prevPID && p.Port == prevPort && p.Protocol == prevProto {
					m.moveCursor(i)
					return m, nil
				}
			}
		}
		m.moveCursor(m.cursor)
		return m, nil

	case killResultMsg:
		if msg.err != nil {
			m.setStatus(fmt.Sprintf("failed to kill PID %d: %v", msg.port.PID, msg.err), 5*time.Second)
		} else if msg.force {
			m.setStatus(fmt.Sprintf("force killed %s (PID %d)", msg.port.Process, msg.port.PID), 3*time.Second)
		} else {
			m.setStatus(fmt.Sprintf("sent SIGTERM to %s (PID %d)", msg.port.Process, msg.port.PID), 3*time.Second)
		}
		m.mode = modeNormal
		return m, loadPorts(m.showSystem)

	case restartResultMsg:
		if msg.err != nil {
			m.setStatus(fmt.Sprintf("restart failed: %v", msg.err), 5*time.Second)
		} else {
			m.setStatus(fmt.Sprintf(
				"restarted %s on %s (PID %d → %d)",
				msg.port.Process, msg.port.DisplayAddress(), msg.port.PID, msg.newPID,
			), 4*time.Second)
		}
		m.mode = modeNormal
		return m, loadPorts(m.showSystem)

	case updateCheckedMsg:
		m.updateInfo = update.Info(msg)
		return m, nil
	}

	return m, nil
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "?":
		m.helpOpen = !m.helpOpen
		return m, nil

	case "U":
		if m.updateInfo.Available {
			m.setStatus("update: "+m.updateInfo.Command, 30*time.Second)
		}
		return m, nil

	case "esc":
		if m.updateInfo.Available && !m.updateDismissed {
			m.updateDismissed = true
			return m, nil
		}
		return m, nil

	case "/":
		m.mode = modeFilter
		m.filterInput.SetValue("")
		m.filterInput.Focus()
		return m, textinput.Blink

	case "r":
		m.setStatus("refreshing...", 1*time.Second)
		return m, loadPorts(m.showSystem)

	case "a", "A":
		m.showSystem = !m.showSystem
		if m.showSystem {
			m.setStatus("showing all tcp listen ports", 2*time.Second)
		} else {
			m.setStatus("showing dev ports only", 2*time.Second)
		}
		return m, loadPorts(m.showSystem)

	case "j", "down":
		m.moveCursor(m.cursor + 1)
		return m, nil

	case "k", "up":
		m.moveCursor(m.cursor - 1)
		return m, nil

	case "g":
		m.moveCursor(0)
		return m, nil

	case "G":
		m.moveCursor(len(m.filtered) - 1)
		return m, nil

	case "d", "x":
		if p := m.selected(); p != nil && p.PID > 0 {
			m.confirmPort = *p
			m.mode = modeConfirmKill
		} else if p := m.selected(); p != nil {
			m.setStatus("no process info (try sudo)", 4*time.Second)
		}
		return m, nil

	case "K":
		if p := m.selected(); p != nil && p.PID > 0 {
			m.confirmPort = *p
			m.mode = modeConfirmForceKill
		}
		return m, nil

	case "s", "S":
		if p := m.selected(); p != nil && p.PID > 0 {
			m.confirmPort = *p
			m.mode = modeConfirmRestart
		} else if p := m.selected(); p != nil {
			m.setStatus("no process info (try sudo)", 4*time.Second)
		}
		return m, nil
	}

	return m, nil
}

func (m Model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.mode = modeNormal
		m.filterInput.Blur()
		m.applyFilter()
		m.moveCursor(m.cursor)
		return m, nil
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	m.applyFilter()
	return m, cmd
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		switch m.mode {
		case modeConfirmRestart:
			return m, restartCmd(m.confirmPort)
		case modeConfirmForceKill:
			return m, killCmd(m.confirmPort, true)
		default:
			return m, killCmd(m.confirmPort, false)
		}

	case "n", "N", "esc":
		m.mode = modeNormal
		m.setStatus("cancelled", 2*time.Second)
		return m, nil
	}
	return m, nil
}

func (m *Model) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(m.filterInput.Value()))
	m.filtered = m.filtered[:0]
	for _, p := range m.allPorts {
		if query == "" {
			m.filtered = append(m.filtered, p)
			continue
		}
		haystack := strings.ToLower(strings.Join([]string{
			p.Protocol, p.State, p.DisplayAddress(), p.Process, p.User, p.Command,
			fmt.Sprintf("%d", p.Port), fmt.Sprintf("%d", p.PID),
		}, " "))
		if strings.Contains(haystack, query) {
			m.filtered = append(m.filtered, p)
		}
	}
	if m.cursor >= len(m.filtered) {
		if len(m.filtered) == 0 {
			m.cursor = 0
		} else {
			m.cursor = len(m.filtered) - 1
		}
	}
	m.syncListOffset()
}

func (m *Model) setStatus(msg string, dur time.Duration) {
	m.statusMsg = msg
	m.statusExpiry = time.Now().Add(dur)
}

func (m Model) selected() *ports.Port {
	if len(m.filtered) == 0 || m.cursor < 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	p := m.filtered[m.cursor]
	return &p
}

func truncate(s string, max int) string {
	if max <= 3 {
		return s
	}
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

const helpText = `lazyports — keyboard shortcuts

  j / k, ↑ / ↓    navigate ports
  g / G           jump to top / bottom
  /               filter
  a               toggle dev / all tcp listen
  d / x           kill (SIGTERM)
  K               force kill (SIGKILL)
  s               restart process
  r               refresh
  U               show update install command
  ?               help
  q               quit`
