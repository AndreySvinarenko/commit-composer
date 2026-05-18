package tui

import tea "github.com/charmbracelet/bubbletea"

// mouseZone tags a region of the screen so click/wheel events can be routed
// to the right action without re-computing layout math everywhere.
type mouseZone int

const (
	zoneNone mouseZone = iota
	zoneLeft
	zoneRight
	zoneStatusBar
	zoneFooter
)

// hitZone classifies a screen coordinate into a layout zone for the main TUI.
// Layout (top to bottom):
//
//	rows 0 .. bodyHeight-1     : panes (pane border drawn at row 0 and bodyHeight-1)
//	row  bodyHeight            : status bar (1 line)
//	rows bodyHeight+1 .. h-1   : help line(s)
func (m Model) hitZone(x, y int) mouseZone {
	if y < 0 || y >= m.height || x < 0 || x >= m.width {
		return zoneNone
	}
	body := m.bodyHeight()
	if y == body {
		return zoneStatusBar
	}
	if y > body {
		return zoneFooter
	}
	left, _ := m.paneWidths()
	if x < left {
		return zoneLeft
	}
	return zoneRight
}

// handleMouse routes a mouse event in the main TUI. Returns the updated model
// + an optional command (typically a diff load on cursor jump).
func (m Model) handleMouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	zone := m.hitZone(msg.X, msg.Y)
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		switch zone {
		case zoneRight:
			m.diff.LineUp(3)
		case zoneLeft:
			if m.cursor > 0 {
				m.cursor--
				m.ensureCursorVisible()
				m.resetDiffViewport()
			}
			return m, m.loadCommitCmd()
		}
	case tea.MouseButtonWheelDown:
		switch zone {
		case zoneRight:
			m.diff.LineDown(3)
		case zoneLeft:
			if m.cursor < len(m.rows)-1 {
				m.cursor++
				m.ensureCursorVisible()
				m.resetDiffViewport()
			}
			return m, m.loadCommitCmd()
		}
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}
		switch zone {
		case zoneLeft:
			m.focus = 0
			// Pane has border row 0, title row 1; content rows start at y=2.
			rowOffset := msg.Y - 2
			if rowOffset >= 0 {
				first, _ := m.listWindow()
				idx := first + rowOffset
				if idx >= 0 && idx < len(m.rows) {
					m.cursor = idx
					m.ensureCursorVisible()
					m.resetDiffViewport()
					return m, m.loadCommitCmd()
				}
			}
		case zoneRight:
			m.focus = 1
		}
	}
	return m, nil
}

// hitZone for the review TUI. Same scheme: panes -> status bar -> footer.
func (m ReviewModel) hitZone(x, y int) mouseZone {
	if y < 0 || y >= m.height || x < 0 || x >= m.width {
		return zoneNone
	}
	body := m.height - 3
	if body < 6 {
		body = 6
	}
	if y == body {
		return zoneStatusBar
	}
	if y > body {
		return zoneFooter
	}
	leftW := m.width * 45 / 100
	if leftW < 30 {
		leftW = 30
	}
	if x < leftW {
		return zoneLeft
	}
	return zoneRight
}

// handleMouse routes a mouse event in the review TUI.
func (m ReviewModel) handleMouse(msg tea.MouseMsg) (ReviewModel, tea.Cmd) {
	zone := m.hitZone(msg.X, msg.Y)
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		switch zone {
		case zoneRight:
			m.diff.LineUp(3)
		case zoneLeft:
			if m.cursor > 0 {
				m.cursor--
			}
			return m, m.maybeLoadDiff()
		}
	case tea.MouseButtonWheelDown:
		switch zone {
		case zoneRight:
			m.diff.LineDown(3)
		case zoneLeft:
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
			return m, m.maybeLoadDiff()
		}
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}
		switch zone {
		case zoneLeft:
			m.focus = 0
			// We render: title (1) + per-pool { blank (1) + header (1) + rule (1) + rows... }.
			// Build a row map so clicks land on the right reviewRow index.
			content := msg.Y - 2 // border + title
			if content < 0 {
				return m, nil
			}
			rowAt := m.rowAtListLine(content)
			if rowAt >= 0 && rowAt < len(m.rows) {
				m.cursor = rowAt
				return m, m.maybeLoadDiff()
			}
		case zoneRight:
			m.focus = 1
		}
	}
	return m, nil
}

// rowAtListLine maps a content-relative line offset in the proposal list to a
// reviewRow index. Returns -1 if the line falls on a header/rule/blank.
//
// Layout repeats per pool: blank, header, rule, then row per group.
func (m ReviewModel) rowAtListLine(line int) int {
	cur := 0
	prevPool := -1
	for i, r := range m.rows {
		if r.poolIdx != prevPool {
			prevPool = r.poolIdx
			cur += 3 // blank + header + rule
		}
		if line == cur {
			return i
		}
		cur++
	}
	return -1
}
