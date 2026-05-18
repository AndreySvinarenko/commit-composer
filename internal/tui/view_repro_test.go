package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// pasteAtBroken is the original (pre-fix) implementation, kept here only to
// pin the bug — it slices `dst` by rune index while treating `col` as a
// visible column, so the cut falls inside an SGR escape and re-glues the
// pieces wrong. TestPasteAtBrokenReproducesBug below proves the old impl
// would still fail today, so TestPasteAtPreservesANSI in view_test.go
// isn't passing by accident.
func pasteAtBroken(dst, src string, col int) string {
	dw := lipgloss.Width(dst)
	if col >= dw {
		return dst + strings.Repeat(" ", col-dw) + src
	}
	rDst := []rune(dst)
	rSrc := []rune(src)
	if col+len(rSrc) > len(rDst) {
		out := append([]rune{}, rDst[:col]...)
		out = append(out, rSrc...)
		return string(out)
	}
	out := append([]rune{}, rDst[:col]...)
	out = append(out, rSrc...)
	out = append(out, rDst[col+len(rSrc):]...)
	return string(out)
}

func TestPasteAtBrokenReproducesBug(t *testing.T) {
	forceTrueColor(t)

	styleA := lipgloss.NewStyle().Foreground(lipgloss.Color("#d58957"))
	styleB := lipgloss.NewStyle().Foreground(lipgloss.Color("#585858"))
	dst := styleA.Render("pick abc1234") + " " + styleB.Render("commit subject")

	const col = 8
	const src = "[OVER]"

	visOld := ansi.Strip(pasteAtBroken(dst, src, col))
	visNew := ansi.Strip(pasteAt(dst, src, col))

	// Old impl: the cut at rune 8 lands inside `\x1b[38;2;213;...m`, so
	// the SGR tail `;87m` reappears as visible text and the total
	// visible width grows past dst's.
	if !strings.Contains(visOld, ";87m") {
		t.Fatalf("old impl should leak SGR tail; got visible %q", visOld)
	}
	if lipgloss.Width(pasteAtBroken(dst, src, col)) == lipgloss.Width(dst) {
		t.Fatalf("old impl should distort visible width; got equal to dst (%d)", lipgloss.Width(dst))
	}

	// New impl: same visible text would appear regardless of style on
	// the underlying row, and width is preserved.
	if visNew != "pick abc[OVER]ommit subject" {
		t.Fatalf("new visible: got %q want %q", visNew, "pick abc[OVER]ommit subject")
	}
	if w := lipgloss.Width(pasteAt(dst, src, col)); w != lipgloss.Width(dst) {
		t.Fatalf("new impl distorted width: got %d want %d", w, lipgloss.Width(dst))
	}
}
