package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mrcat71/commit-composer/internal/git"
	"github.com/mrcat71/commit-composer/internal/tui"
)

func main() {
	commits := make([]git.Commit, 25)
	for i := range commits {
		commits[i] = git.Commit{
			SHA:     fmt.Sprintf("%040d", i),
			Short:   fmt.Sprintf("abc%04d", i),
			Author:  "Andrey",
			Email:   "andrei@example",
			Date:    time.Date(2026, 5, 12, 17, 36, 5, 0, time.UTC),
			Subject: fmt.Sprintf("commit %d title here", i),
		}
	}
	m := tui.New(tui.Options{Commits: commits, Base: "base", RangeSpec: "base..HEAD",
		LoadFiles: func(string) ([]git.FileStat, error) {
			return []git.FileStat{
				{Status: "M", Path: "dot_agent-skills/helm-worker/SKILL.md"},
			}, nil
		}})
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	out := mm.View().Content
	// Strip ANSI for legibility.
	stripped := strip(out)
	lines := strings.Split(stripped, "\n")
	for i, l := range lines {
		fmt.Printf("%3d | %s\n", i+1, l)
	}
}

func strip(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		if r == 0x1b {
			in = true
			continue
		}
		if in {
			if r == 'm' || r == 'K' {
				in = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
