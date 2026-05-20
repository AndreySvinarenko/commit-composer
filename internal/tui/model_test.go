package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mrcat71/commit-composer/internal/git"
	"github.com/mrcat71/commit-composer/internal/plan"
)

// keyPress is a small helper that drives one key through the TUI's Update.
// Most tests just want the resulting Model; the returned Cmd is ignored
// unless the test specifically asserts on it.
func keyPress(m Model, s string) Model {
	msg := tea.KeyPressMsg{Text: s}
	out, _ := m.Update(msg)
	mm, _ := out.(Model)
	return mm
}

func sampleCommits(n int) []git.Commit {
	out := make([]git.Commit, n)
	for i := 0; i < n; i++ {
		sha := "0000000000000000000000000000000000000000"
		out[i] = git.Commit{
			SHA:     sha[:40-len(itoa(i))] + itoa(i),
			Short:   "abc" + itoa(i),
			Subject: "commit " + itoa(i),
		}
	}
	return out
}

func itoa(i int) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{digits[i%10]}, b...)
		i /= 10
	}
	return string(b)
}

func TestToggleRecompose(t *testing.T) {
	m := New(Options{Commits: sampleCommits(2)})
	// First press: Pick -> ClaudeRecompose.
	m = m.toggleRecompose()
	if m.rows[0].action != plan.ClaudeRecompose {
		t.Fatalf("action: got %v want ClaudeRecompose", m.rows[0].action)
	}
	// Second press: ClaudeRecompose -> Pick (toggle off).
	m = m.toggleRecompose()
	if m.rows[0].action != plan.Pick {
		t.Errorf("after toggling off: got %v want Pick", m.rows[0].action)
	}
}

func TestPoolBoundsConsecutive(t *testing.T) {
	m := New(Options{Commits: sampleCommits(5)})
	// Mark rows 1, 2, 3 as recompose; rows 0 and 4 stay pick.
	for _, i := range []int{1, 2, 3} {
		m.cursor = i
		m = m.toggleRecompose()
	}
	first, last := m.poolBounds(2)
	if first != 1 || last != 4 {
		t.Errorf("poolBounds(2) = (%d, %d), want (1, 4)", first, last)
	}
	// A non-marked row reports a 1-wide pool.
	first, last = m.poolBounds(0)
	if first != 0 || last != 0 {
		t.Errorf("poolBounds(0) on unmarked = (%d, %d), want (0, 0)", first, last)
	}
}

func TestPlanReflectsRows(t *testing.T) {
	m := New(Options{
		Commits:   sampleCommits(3),
		Base:      "deadbeef",
		RangeSpec: "deadbeef..HEAD",
	})
	// Mutate the action on the middle row.
	m.cursor = 1
	m = m.setAction(plan.Squash)
	p := m.Plan()
	if p.Base != "deadbeef" || p.Range != "deadbeef..HEAD" {
		t.Fatalf("plan metadata mismatch: %+v", p)
	}
	if len(p.Ops) != 3 {
		t.Fatalf("ops len: %d", len(p.Ops))
	}
	wants := []plan.Action{plan.Pick, plan.Squash, plan.Pick}
	for i, w := range wants {
		if p.Ops[i].Action != w {
			t.Errorf("op[%d]: got %v want %v", i, p.Ops[i].Action, w)
		}
		if p.Ops[i].OrigIndex != i {
			t.Errorf("op[%d]: OrigIndex %d want %d", i, p.Ops[i].OrigIndex, i)
		}
	}
}

func TestValidate(t *testing.T) {
	mkRows := func(actions ...plan.Action) []row {
		rs := make([]row, len(actions))
		for i, a := range actions {
			rs[i] = row{
				commit: git.Commit{Short: "abc" + itoa(i)},
				action: a,
			}
		}
		return rs
	}
	// Rows are newest-first: index 0 = newest, last index = oldest.
	// The rebase applies them in REVERSE (oldest first). validate() checks
	// the "earliest applied" = last non-dropped row in display order.
	tests := []struct {
		name string
		rows []row
		ok   bool
	}{
		{"all pick", mkRows(plan.Pick, plan.Pick), true},
		// Oldest (last row) is Squash with nothing before it - invalid.
		{"oldest squash invalid", mkRows(plan.Pick, plan.Squash), false},
		{"oldest fixup invalid", mkRows(plan.Pick, plan.Fixup), false},
		// Drop the oldest; earliest applied becomes the next-oldest Pick - ok.
		{"drop oldest is ok", mkRows(plan.Pick, plan.Drop), true},
		{"all dropped invalid", mkRows(plan.Drop, plan.Drop), false},
		// Drop oldest, but next is Squash - first kept is Squash - invalid.
		{"drop oldest then squash next is invalid", mkRows(plan.Pick, plan.Squash, plan.Drop), false},
		// Mixed: as long as the earliest applied is Pick (or any non-squash/fixup), ok.
		{"mixed ok (earliest is Pick)", mkRows(plan.Drop, plan.Edit, plan.Reword, plan.Fixup, plan.Squash, plan.Pick), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(tc.rows)
			if tc.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestRewordChooserOpensOnR(t *testing.T) {
	m := New(Options{Commits: sampleCommits(2)})
	m = keyPress(m, "r")
	if !m.rewordChooser {
		t.Fatal("expected rewordChooser=true after pressing r")
	}
	if m.status == "" {
		t.Error("expected non-empty status hint while chooser is open")
	}
	if m.rows[m.cursor].action != plan.Pick {
		t.Errorf("row action should still be Pick (chooser pending), got %v", m.rows[m.cursor].action)
	}
}

func TestRewordChooserCMarksClaudeReword(t *testing.T) {
	m := New(Options{Commits: sampleCommits(2)})
	m = keyPress(m, "r")
	m = keyPress(m, "c")
	if m.rewordChooser {
		t.Fatal("chooser should be closed after pressing c")
	}
	if m.rows[m.cursor].action != plan.ClaudeReword {
		t.Errorf("row action: got %v want ClaudeReword", m.rows[m.cursor].action)
	}
	if m.rows[m.cursor].reword != "" {
		t.Errorf("claude-reword path must not pre-fill reword; got %q", m.rows[m.cursor].reword)
	}
	if m.status == "" {
		t.Error("expected confirmation status after claude-reword mark")
	}
}

func TestRewordChooserQClosesWithoutChange(t *testing.T) {
	m := New(Options{Commits: sampleCommits(2)})
	m = keyPress(m, "r")
	m = keyPress(m, "q") // also covers the esc branch
	if m.rewordChooser {
		t.Fatal("chooser should be closed after pressing q/esc")
	}
	if m.rows[m.cursor].action != plan.Pick {
		t.Errorf("row action should stay Pick after cancelling chooser, got %v", m.rows[m.cursor].action)
	}
	if m.status != "" {
		t.Errorf("status should be cleared after cancel, got %q", m.status)
	}
}

func TestPlanEmitsClaudeReword(t *testing.T) {
	m := New(Options{
		Commits:   sampleCommits(2),
		Base:      "deadbeef",
		RangeSpec: "deadbeef..HEAD",
	})
	m = keyPress(m, "r")
	m = keyPress(m, "c")
	p := m.Plan()
	var found bool
	for _, op := range p.Ops {
		if op.Action == plan.ClaudeReword {
			found = true
			if op.NewMessage != "" {
				t.Errorf("claude-reword op should not carry NewMessage from the TUI; got %q", op.NewMessage)
			}
		}
	}
	if !found {
		t.Fatal("Plan() did not surface the claude-reword action")
	}
}

func TestReorderUpdatesPlan(t *testing.T) {
	m := New(Options{Commits: sampleCommits(3)})
	// sampleCommits(3) returns [commit 0, commit 1, commit 2] in oldest-first
	// order. After tui.New reverses for display, m.rows is newest-first:
	//   rows[0] = commit 2 (OrigIndex=2)
	//   rows[1] = commit 1 (OrigIndex=1)
	//   rows[2] = commit 0 (OrigIndex=0)
	// Plan() reverses back to ops[0..2] = OrigIndex 0,1,2 -> not reordered.
	if p := m.Plan(); p.Reordered() {
		t.Fatalf("initial plan should not be reordered; ops=%+v", p.Ops)
	}
	// Swap rows[0] and rows[1] in the display - now ops differ from OrigIndex.
	m.rows[0], m.rows[1] = m.rows[1], m.rows[0]
	p := m.Plan()
	if !p.Reordered() {
		t.Fatalf("after row swap, Reordered should be true; ops=%+v", p.Ops)
	}
}
