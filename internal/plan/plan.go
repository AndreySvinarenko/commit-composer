// Package plan defines the data model for a commit-composer rebase plan:
// the user's marked-up actions over a contiguous slice of commits.
package plan

import "fmt"

// Action is a per-commit decision the user has made in the TUI.
type Action int

const (
	Pick Action = iota
	Reword
	Squash
	Fixup
	Drop
	Edit
	// ClaudeRecompose marks the commit for Claude-driven redesign.
	//
	// Consecutive ClaudeRecompose commits are pooled into a single
	// analysis batch: their changes are unioned and Claude proposes a
	// fresh sequence of commits (typically grouped by feature/topic).
	//
	// The proposal is reviewed (and optionally edited) before any
	// rebase happens.
	ClaudeRecompose
	// ClaudeReword marks the commit for Claude-assisted rewording.
	//
	// The TUI emits these as bare `claude-reword <sha>` op lines (no
	// trailing message). The slash command resolves them before the
	// rebase: Claude proposes a new message from the current message +
	// diff, the user reviews/edits in $EDITOR per commit, and a binary
	// helper rewrites each ClaudeReword op into a regular Reword op
	// carrying the accepted message. The apply layer therefore never
	// sees ClaudeReword - it only ever sees Reword.
	ClaudeReword
)

// NumActions reports the count of valid Action values. The TUI uses this to
// drive the Space-cycle wraparound; keeping it here lets the enum stay the
// single source of truth.
func NumActions() int { return 8 }

// String returns the lowercase token used in the serialized plan.
func (a Action) String() string {
	switch a {
	case Pick:
		return "pick"
	case Reword:
		return "reword"
	case Squash:
		return "squash"
	case Fixup:
		return "fixup"
	case Drop:
		return "drop"
	case Edit:
		return "edit"
	case ClaudeRecompose:
		return "claude-recompose"
	case ClaudeReword:
		return "claude-reword"
	default:
		return fmt.Sprintf("action(%d)", int(a))
	}
}

// ParseAction converts a token from the serialized plan back into an Action.
func ParseAction(s string) (Action, error) {
	switch s {
	case "pick":
		return Pick, nil
	case "reword":
		return Reword, nil
	case "squash":
		return Squash, nil
	case "fixup":
		return Fixup, nil
	case "drop":
		return Drop, nil
	case "edit":
		return Edit, nil
	case "claude-recompose":
		return ClaudeRecompose, nil
	// Backward-compat with the previous "claude-split" token.
	case "claude-split":
		return ClaudeRecompose, nil
	case "claude-reword":
		return ClaudeReword, nil
	default:
		return 0, fmt.Errorf("unknown action %q", s)
	}
}

// Op is one entry in a Plan: an action targeting a specific commit.
//
// OrigIndex records the commit's original position in the input range so that
// callers can detect reorderings (Ops slice reordered vs OrigIndex monotonic).
type Op struct {
	SHA        string
	Action     Action
	NewMessage string // for Reword
	OrigIndex  int
}

// Plan is the full proposal emitted by the TUI on Enter.
//
// Base is the immutable parent commit; the rebase will target Base^..Head.
// Ops are stored in the order they should appear in the final history.
type Plan struct {
	Base  string
	Range string
	Ops   []Op
}

// Reordered reports whether the Ops slice no longer matches OrigIndex order.
func (p Plan) Reordered() bool {
	for i, op := range p.Ops {
		if op.OrigIndex != i {
			return true
		}
	}
	return false
}
