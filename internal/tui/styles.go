package tui

import "github.com/charmbracelet/lipgloss"

// Lazydocker-inspired palette (default theme from lazydocker config).
const (
	colorBorderActive   = "2"  // green
	colorBorderInactive = "8"  // dark grey / default-ish
	colorSelectedBg     = "4"  // blue
	colorSelectedFg     = "15" // white on blue selection
	colorOptions        = "4"  // blue key hints
	colorStatus         = "6"  // cyan
	colorInfo           = "2"  // green
	colorWarn           = "3"  // yellow
	colorError          = "1"  // red
	colorMagenta        = "5"  // process names
	colorYellow         = "3"  // port numbers
	colorDim            = "8"  // muted text
)

var (
	appStyle = lipgloss.NewStyle().
			Align(lipgloss.Left, lipgloss.Top)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorInfo))

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorBorderActive))

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(colorBorderActive))

	inactivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(colorBorderInactive))

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(colorDim))

	rowStyle = lipgloss.NewStyle()

	selectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(colorSelectedBg)).
				Foreground(lipgloss.Color(colorSelectedFg))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim))

	statusBarStyle = lipgloss.NewStyle().
			Padding(0, 1)

	statusAccentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorStatus)).
				Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorError)).
			Bold(true)

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorWarn)).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorInfo))

	optionsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorOptions))

	portStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorYellow))

	processStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMagenta))

	filterPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorInfo))

	filterInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorInfo))

	helpOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(colorBorderActive)).
				Padding(1, 2)
)
