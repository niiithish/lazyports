package tui

import "github.com/charmbracelet/lipgloss"

const (
	sidePanelRatio = 0.3333 // lazydocker default sidePanelWidth
	statusRows     = 2
)

type layout struct {
	width       int
	height      int
	leftInnerW  int // content width inside left panel border
	rightInnerW int // content width inside right panel border
	bodyH       int // total body height including panel borders
	bodyInnerH  int // content height inside panel borders
	listH       int
	detailH     int
}

func computeLayout(width, height int) layout {
	l := layout{width: width, height: height}

	// Lipgloss Width/Height apply to content only; each bordered panel adds 2 cols/lines.
	contentW := width - 4
	if contentW < 40 {
		contentW = 40
	}

	l.leftInnerW = int(float64(contentW) * sidePanelRatio)
	if l.leftInnerW < 22 {
		l.leftInnerW = 22
	}
	l.rightInnerW = contentW - l.leftInnerW
	if l.rightInnerW < 18 {
		l.rightInnerW = 18
		l.leftInnerW = contentW - l.rightInnerW
	}

	l.bodyH = height - statusRows
	if l.bodyH < 6 {
		l.bodyH = 6
	}

	l.bodyInnerH = l.bodyH - 2 // top + bottom border
	if l.bodyInnerH < 4 {
		l.bodyInnerH = 4
	}

	// title + header + divider + rows
	l.listH = l.bodyInnerH - 3
	if l.listH < 2 {
		l.listH = 2
	}

	// title + detail lines
	l.detailH = l.bodyInnerH - 1
	if l.detailH < 3 {
		l.detailH = 3
	}

	return l
}

func protoStyle(proto string) lipgloss.Style {
	if proto == "TCP" {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorInfo))
	}
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorDim))
}

func fitLines(lines []string, height int) []string {
	if height < 1 {
		height = 1
	}
	out := make([]string, height)
	for i := 0; i < height; i++ {
		if i < len(lines) {
			out[i] = lines[i]
		}
	}
	return out
}
