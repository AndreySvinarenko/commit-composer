package plan

import (
	"strings"
	"testing"
)

func TestActionStringRoundTrip(t *testing.T) {
	cases := []Action{Pick, Reword, Squash, Fixup, Drop, Edit, ClaudeRecompose, ClaudeReword}
	for _, a := range cases {
		t.Run(a.String(), func(t *testing.T) {
			got, err := ParseAction(a.String())
			if err != nil {
				t.Fatalf("ParseAction(%q) err: %v", a.String(), err)
			}
			if got != a {
				t.Fatalf("round-trip mismatch: %v -> %q -> %v", a, a.String(), got)
			}
		})
	}
}

func TestParseActionUnknown(t *testing.T) {
	if _, err := ParseAction("bogus"); err == nil {
		t.Fatal("expected error for unknown action")
	}
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		in   Plan
	}{
		{
			name: "simple pick chain",
			in: Plan{
				Base:  "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
				Range: "deadbeef..HEAD",
				Ops: []Op{
					{SHA: "a1b2c3d", Action: Pick, OrigIndex: 0},
					{SHA: "e4f5a6b", Action: Pick, OrigIndex: 1},
				},
			},
		},
		{
			name: "mixed actions with single-line reword",
			in: Plan{
				Base: "0000000000000000000000000000000000000000",
				Ops: []Op{
					{SHA: "aaa", Action: Pick, OrigIndex: 0},
					{SHA: "bbb", Action: Reword, NewMessage: "fix: tighten auth", OrigIndex: 1},
					{SHA: "ccc", Action: Squash, OrigIndex: 2},
					{SHA: "ddd", Action: Fixup, OrigIndex: 3},
					{SHA: "eee", Action: Drop, OrigIndex: 4},
					{SHA: "fff", Action: Edit, OrigIndex: 5},
				},
			},
		},
		{
			name: "multiline reword via base64",
			in: Plan{
				Base: "0000000000000000000000000000000000000000",
				Ops: []Op{
					{
						SHA:        "abc1234",
						Action:     Reword,
						NewMessage: "feat: new thing\n\nWith body paragraph.\n\nAnd a second one.",
						OrigIndex:  0,
					},
				},
			},
		},
		{
			name: "reordered ops",
			in: Plan{
				Base: "0000000000000000000000000000000000000000",
				Ops: []Op{
					{SHA: "two", Action: Pick, OrigIndex: 1},
					{SHA: "one", Action: Pick, OrigIndex: 0},
					{SHA: "three", Action: Squash, OrigIndex: 2},
				},
			},
		},
		{
			name: "claude-recompose ops",
			in: Plan{
				Base: "0000000000000000000000000000000000000000",
				Ops: []Op{
					{SHA: "aaa", Action: ClaudeRecompose, OrigIndex: 0},
					{SHA: "bbb", Action: ClaudeRecompose, OrigIndex: 1},
					{SHA: "ccc", Action: ClaudeRecompose, OrigIndex: 2},
				},
			},
		},
		{
			name: "claude-reword ops (no message attached)",
			in: Plan{
				Base: "0000000000000000000000000000000000000000",
				Ops: []Op{
					{SHA: "aaa", Action: Pick, OrigIndex: 0},
					{SHA: "bbb", Action: ClaudeReword, OrigIndex: 1},
					{SHA: "ccc", Action: ClaudeReword, OrigIndex: 2},
				},
			},
		},
		{
			name: "claude-reword op with pre-filled message (round-trips through b64 if multiline)",
			in: Plan{
				Base: "0000000000000000000000000000000000000000",
				Ops: []Op{
					{
						SHA:        "abc1234",
						Action:     ClaudeReword,
						NewMessage: "feat(scope): proposed line\n\nWith body.",
						OrigIndex:  0,
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := Marshal(tc.in)
			got, err := Unmarshal(strings.NewReader(s))
			if err != nil {
				t.Fatalf("Unmarshal err: %v\n--- payload ---\n%s", err, s)
			}
			if got.Base != tc.in.Base {
				t.Errorf("Base: got %q want %q", got.Base, tc.in.Base)
			}
			if got.Range != tc.in.Range {
				t.Errorf("Range: got %q want %q", got.Range, tc.in.Range)
			}
			if len(got.Ops) != len(tc.in.Ops) {
				t.Fatalf("Ops len: got %d want %d", len(got.Ops), len(tc.in.Ops))
			}
			for i := range got.Ops {
				want := tc.in.Ops[i]
				g := got.Ops[i]
				// OrigIndex on unmarshal reflects position in the serialized
				// stream, which may not match the input's original index when
				// the input was reordered. Compare the other fields only.
				if g.SHA != want.SHA || g.Action != want.Action || g.NewMessage != want.NewMessage {
					t.Errorf("op[%d]: got %+v want %+v", i, g, want)
				}
			}
		})
	}
}

func TestUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"missing header", "base: abc\nops:\n- pick aaa\n"},
		{"missing base", "## commit-composer plan v1\nops:\n- pick aaa\n"},
		{"unknown action", "## commit-composer plan v1\nbase: deadbeef\nops:\n- nuke aaa\n"},
		{"op missing sha", "## commit-composer plan v1\nbase: deadbeef\nops:\n- pick\n"},
		{"bad b64 reword", "## commit-composer plan v1\nbase: deadbeef\nops:\n- reword aaa :: b64::!!notbase64!!\n"},
		{"tail on action that does not accept it", "## commit-composer plan v1\nbase: deadbeef\nops:\n- pick aaa :: junk\n"},
		{"op line not prefixed", "## commit-composer plan v1\nbase: deadbeef\nops:\npick aaa\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Unmarshal(strings.NewReader(tc.in)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestUnmarshalToleratesTrailingWhitespace(t *testing.T) {
	in := "## commit-composer plan v1   \nbase: deadbeef\t\nops:\n- pick aaa  \n- pick bbb\t\r\n"
	p, err := Unmarshal(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Unmarshal err: %v", err)
	}
	if p.Base != "deadbeef" {
		t.Errorf("Base: got %q want deadbeef", p.Base)
	}
	if len(p.Ops) != 2 {
		t.Fatalf("ops len: got %d want 2", len(p.Ops))
	}
}

func TestReordered(t *testing.T) {
	tests := []struct {
		name string
		ops  []Op
		want bool
	}{
		{
			name: "monotonic",
			ops: []Op{
				{OrigIndex: 0}, {OrigIndex: 1}, {OrigIndex: 2},
			},
			want: false,
		},
		{
			name: "swapped",
			ops: []Op{
				{OrigIndex: 1}, {OrigIndex: 0}, {OrigIndex: 2},
			},
			want: true,
		},
		{
			name: "drop in middle still monotonic",
			ops: []Op{
				{OrigIndex: 0}, {OrigIndex: 1}, {OrigIndex: 2},
			},
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Plan{Ops: tc.ops}
			if got := p.Reordered(); got != tc.want {
				t.Errorf("Reordered: got %v want %v", got, tc.want)
			}
		})
	}
}
