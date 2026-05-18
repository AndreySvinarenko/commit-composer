package tui

import "github.com/charmbracelet/lipgloss"

// styles bundles all lipgloss styles used by the TUI. Centralized so theming
// changes are a single edit. All values derive from the Theme palette.
type styles struct {
	theme Theme

	pane          lipgloss.Style
	paneFocused   lipgloss.Style
	title         lipgloss.Style
	cursor        lipgloss.Style
	row           lipgloss.Style
	rowSelected   lipgloss.Style
	tag           lipgloss.Style
	tagPick       lipgloss.Style
	tagReword     lipgloss.Style
	tagSquash     lipgloss.Style
	tagFixup      lipgloss.Style
	tagDrop       lipgloss.Style
	tagEdit       lipgloss.Style
	tagRecompose  lipgloss.Style
	short         lipgloss.Style
	subject       lipgloss.Style
	subjectMuted  lipgloss.Style
	help          lipgloss.Style
	helpKey       lipgloss.Style
	status        lipgloss.Style
	statusError   lipgloss.Style
	statusSuccess lipgloss.Style
	statusBar     lipgloss.Style
	statusBarKey  lipgloss.Style
	diffAdd       lipgloss.Style
	diffDel       lipgloss.Style
	diffHunk      lipgloss.Style
	diffFile      lipgloss.Style
	diffContext   lipgloss.Style
	gutter        lipgloss.Style
	scrollbar     lipgloss.Style
	scrollbarBg   lipgloss.Style
	meta          lipgloss.Style
	metaKey       lipgloss.Style
	modal         lipgloss.Style
	modalTitle    lipgloss.Style
}

func newStyles() styles {
	return newStylesFromTheme(defaultTheme())
}

func newStylesFromTheme(t Theme) styles {
	border := lipgloss.RoundedBorder()
	color := func(hex string) lipgloss.Color { return lipgloss.Color(hex) }
	return styles{
		theme: t,
		pane: lipgloss.NewStyle().
			Border(border).
			BorderForeground(color(t.Border)).
			Padding(0, 1),
		paneFocused: lipgloss.NewStyle().
			Border(border).
			BorderForeground(color(t.Accent)).
			Padding(0, 1),
		title: lipgloss.NewStyle().
			Foreground(color(t.Accent)).
			Bold(true),
		cursor: lipgloss.NewStyle().
			Foreground(color(t.CursorFg)).
			Bold(true),
		row: lipgloss.NewStyle().Foreground(color(t.Normal)),
		rowSelected: lipgloss.NewStyle().
			Background(color(t.SelectedBg)).
			Foreground(color(t.SelectedFg)).
			Bold(true),
		tag: lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true),
		// pick = default unmarked state: muted so marked commits stand out.
		tagPick: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(color(t.Muted)),
		tagReword: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("#5f87d7")).
			Foreground(lipgloss.Color("#1a1a1a")),
		tagSquash: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("#D5895F")).
			Foreground(lipgloss.Color("#1a1a1a")),
		tagFixup: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("#aa6d3f")).
			Foreground(lipgloss.Color("#1a1a1a")),
		tagDrop: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("#ff5f5f")).
			Foreground(lipgloss.Color("#1a1a1a")),
		tagEdit: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("#af87d7")).
			Foreground(lipgloss.Color("#1a1a1a")),
		tagRecompose: lipgloss.NewStyle().
			Padding(0, 1).
			Background(color(t.WarnFg)).
			Foreground(lipgloss.Color("#1a1a1a")).
			Bold(true),
		short: lipgloss.NewStyle().
			Foreground(color(t.Muted)),
		subject: lipgloss.NewStyle().
			Foreground(color(t.Normal)),
		subjectMuted: lipgloss.NewStyle().
			Foreground(color(t.Muted)),
		help: lipgloss.NewStyle().
			Foreground(color(t.Muted)),
		helpKey: lipgloss.NewStyle().
			Foreground(color(t.Accent)).
			Bold(true),
		status: lipgloss.NewStyle().
			Foreground(color(t.Muted)),
		statusError: lipgloss.NewStyle().
			Foreground(color(t.ErrorFg)).
			Bold(true),
		statusSuccess: lipgloss.NewStyle().
			Foreground(color(t.SuccessFg)).
			Bold(true),
		statusBar: lipgloss.NewStyle().
			Foreground(color(t.StatusFg)).
			Background(color(t.StatusBg)).
			Padding(0, 1),
		statusBarKey: lipgloss.NewStyle().
			Foreground(color(t.StatusFg)).
			Background(color(t.StatusBg)).
			Bold(true),
		diffAdd: lipgloss.NewStyle().
			Foreground(color(t.AddFg)),
		diffDel: lipgloss.NewStyle().
			Foreground(color(t.RemoveFg)),
		diffHunk: lipgloss.NewStyle().
			Foreground(color(t.HunkFg)).
			Bold(true),
		diffFile: lipgloss.NewStyle().
			Foreground(color(t.FileHeaderFg)).
			Bold(true),
		diffContext: lipgloss.NewStyle().
			Foreground(color(t.Normal)),
		gutter: lipgloss.NewStyle().
			Foreground(color(t.Muted)),
		scrollbar: lipgloss.NewStyle().
			Foreground(color(t.Accent)),
		scrollbarBg: lipgloss.NewStyle().
			Foreground(color(t.Border)),
		meta: lipgloss.NewStyle().
			Foreground(color(t.Muted)),
		metaKey: lipgloss.NewStyle().
			Foreground(color(t.Normal)).
			Bold(true),
		modal: lipgloss.NewStyle().
			Border(border).
			BorderForeground(color(t.Accent)).
			Padding(1, 2),
		modalTitle: lipgloss.NewStyle().
			Foreground(color(t.Accent)).
			Bold(true),
	}
}
