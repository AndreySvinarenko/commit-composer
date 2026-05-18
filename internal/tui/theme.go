package tui

// Theme bundles the named color tokens used across the TUI. Colors are hex
// strings ("#RRGGBB"); lipgloss handles 24-bit truecolor terminals natively
// and falls back to 256-color or 16-color approximations elsewhere.
//
// Defaults match revdiff's warm-orange palette so the two tools share a
// visual identity when reviewing diffs side by side.
type Theme struct {
	Accent       string // active pane borders, titles, key hints
	Border       string // inactive pane borders, separators
	Normal       string // body text, context lines
	Muted        string // dates, file paths, secondary info, line numbers
	SelectedFg   string // foreground for the selected row
	SelectedBg   string // background for the selected row
	CursorFg     string // cursor glyph
	AddFg        string // diff added line foreground
	AddBg        string // diff added line background
	RemoveFg     string // diff removed line foreground
	RemoveBg     string // diff removed line background
	HunkFg       string // diff @@-hunk header
	FileHeaderFg string // diff +++/--- file path header
	StatusFg     string // status bar / footer foreground
	StatusBg     string // status bar / footer background
	WarnFg       string // amber / warning
	ErrorFg      string // hard error
	SuccessFg    string // success
}

// defaultTheme returns the baked-in revdiff-derived theme.
func defaultTheme() Theme {
	return Theme{
		Accent:       "#D5895F",
		Border:       "#585858",
		Normal:       "#d0d0d0",
		Muted:        "#9e9e9e",
		SelectedFg:   "#ffffaf",
		SelectedBg:   "#D5895F",
		CursorFg:     "#bbbb44",
		AddFg:        "#87d787",
		AddBg:        "#123800",
		RemoveFg:     "#ff8787",
		RemoveBg:     "#4D1100",
		HunkFg:       "#5fafd7",
		FileHeaderFg: "#D5895F",
		StatusFg:     "#202020",
		StatusBg:     "#C5794F",
		WarnFg:       "#f5c542",
		ErrorFg:      "#ff5f5f",
		SuccessFg:    "#87d787",
	}
}
