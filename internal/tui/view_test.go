package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

// forceTrueColor forces lipgloss to emit SGR codes during tests. Without
// it, lipgloss detects a non-TTY stdout and strips all styling, which
// hides the very ANSI-handling behavior we want to exercise.
func forceTrueColor(t *testing.T) {
	t.Helper()
	prev := lipgloss.DefaultRenderer().ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
}

// TestPasteAtPreservesANSI ensures pasteAt never cuts through an ANSI
// escape sequence in the destination. The previous implementation sliced
// `dst` by rune index while treating `col` as a visible column, so when
// the help modal landed on a styled left-pane row the cut fell inside a
// `\x1b[38;2;R;G;Bm` sequence and the tail (`;2;...;m`) printed as text.
func TestPasteAtPreservesANSI(t *testing.T) {
	forceTrueColor(t)
	// Mimic a styled left pane row: a true-color SGR wrapping some text,
	// then more styled text after it.
	styleA := lipgloss.NewStyle().Foreground(lipgloss.Color("#d58957"))
	styleB := lipgloss.NewStyle().Foreground(lipgloss.Color("#585858"))
	dst := styleA.Render("pick abc1234") + " " + styleB.Render("commit subject")

	src := "[OVERLAY]"
	out := pasteAt(dst, src, 8)

	visible := ansi.Strip(out)

	// The leaked tail from the previous bug looked like `;2;R;G;Bm`. If
	// even one ANSI escape sequence got broken open, its tail would now
	// be sitting in the visible text — assert otherwise.
	for _, leak := range []string{";2;213;137;95m", ";2;88;88;88m", "[38;", "[48;"} {
		if strings.Contains(visible, leak) {
			t.Fatalf("ANSI fragment leaked into visible output: found %q in %q", leak, visible)
		}
	}

	// The overlay must appear in the visible text where we asked.
	if !strings.Contains(visible, "[OVERLAY]") {
		t.Fatalf("overlay missing from visible output: %q", visible)
	}

	// Width should still match the dst width since src fits inside it.
	dstW := lipgloss.Width(dst)
	outW := lipgloss.Width(out)
	if outW != dstW {
		t.Fatalf("width changed: dst=%d out=%d (visible=%q)", dstW, outW, visible)
	}
}

// TestPasteAtPastEnd handles the edge where col is past dst's visible
// width — we must pad with spaces, not panic or smear bytes.
func TestPasteAtPastEnd(t *testing.T) {
	forceTrueColor(t)
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#d58957"))
	dst := style.Render("short")
	src := "X"

	out := pasteAt(dst, src, 10)
	visible := ansi.Strip(out)

	if !strings.HasSuffix(visible, "X") {
		t.Fatalf("overlay should land at the end: %q", visible)
	}
	// Visible width should be exactly col + width(src) = 11.
	if w := lipgloss.Width(out); w != 11 {
		t.Fatalf("width: got %d want 11 (visible=%q)", w, visible)
	}
}

// TestPasteAtFullCover replaces the whole line — exercises the
// "col+sw >= dw" branch where there's no right-side remainder.
func TestPasteAtFullCover(t *testing.T) {
	forceTrueColor(t)
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#d58957"))
	dst := style.Render("ABCDE")
	src := "12345"

	out := pasteAt(dst, src, 0)
	visible := ansi.Strip(out)

	if visible != "12345" {
		t.Fatalf("visible: got %q want %q", visible, "12345")
	}
	for _, leak := range []string{"[38;", "[48;", ";2;"} {
		if strings.Contains(visible, leak) {
			t.Fatalf("ANSI fragment leaked: %q in %q", leak, visible)
		}
	}
}
