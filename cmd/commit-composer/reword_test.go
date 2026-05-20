package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrcat71/commit-composer/internal/plan"
)

func writePlan(t *testing.T, dir string, p plan.Plan) string {
	t.Helper()
	path := filepath.Join(dir, "plan.txt")
	if err := os.WriteFile(path, []byte(plan.Marshal(p)), 0o600); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	return path
}

func TestRewordApplyRewritesPlan(t *testing.T) {
	tmp := t.TempDir()
	rewords := filepath.Join(tmp, "rewords")
	if err := os.MkdirAll(rewords, 0o700); err != nil {
		t.Fatalf("mkdir rewords: %v", err)
	}

	sha := "aaaaaaa0000000000000000000000000000000aa"
	planPath := writePlan(t, tmp, plan.Plan{
		Base: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		Ops: []plan.Op{
			{SHA: "bbb", Action: plan.Pick, OrigIndex: 0},
			{SHA: sha, Action: plan.ClaudeReword, OrigIndex: 1},
		},
	})

	finalMsg := "feat(scope): hand-written final message"
	finalPath := filepath.Join(rewords, sha+".reword.final.txt")
	if err := os.WriteFile(finalPath, []byte(finalMsg+"\n"), 0o600); err != nil {
		t.Fatalf("write final: %v", err)
	}

	if err := rewordApplyMain([]string{"--plan=" + planPath, "--rewords-dir=" + rewords}); err != nil {
		t.Fatalf("rewordApplyMain: %v", err)
	}

	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("read rewritten plan: %v", err)
	}
	got := string(data)
	if strings.Contains(got, "claude-reword") {
		t.Errorf("rewritten plan still contains claude-reword:\n%s", got)
	}
	want := "- reword " + sha + " :: " + finalMsg
	if !strings.Contains(got, want) {
		t.Errorf("rewritten plan missing %q\nfull:\n%s", want, got)
	}
}

func TestRewordApplyMissingFinalErrors(t *testing.T) {
	tmp := t.TempDir()
	rewords := filepath.Join(tmp, "rewords")
	if err := os.MkdirAll(rewords, 0o700); err != nil {
		t.Fatalf("mkdir rewords: %v", err)
	}

	sha := "aaaaaaa0000000000000000000000000000000aa"
	planPath := writePlan(t, tmp, plan.Plan{
		Base: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		Ops: []plan.Op{
			{SHA: sha, Action: plan.ClaudeReword, OrigIndex: 0},
		},
	})

	err := rewordApplyMain([]string{"--plan=" + planPath, "--rewords-dir=" + rewords})
	if err == nil {
		t.Fatal("expected error when .reword.final.txt is missing, got nil")
	}
	if !strings.Contains(err.Error(), sha) {
		t.Errorf("error should name the missing sha %s, got: %v", sha, err)
	}
}

func TestRewordApplyEmptyFinalErrors(t *testing.T) {
	tmp := t.TempDir()
	rewords := filepath.Join(tmp, "rewords")
	if err := os.MkdirAll(rewords, 0o700); err != nil {
		t.Fatalf("mkdir rewords: %v", err)
	}

	sha := "bbbbbbb0000000000000000000000000000000bb"
	planPath := writePlan(t, tmp, plan.Plan{
		Base: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		Ops: []plan.Op{
			{SHA: sha, Action: plan.ClaudeReword, OrigIndex: 0},
		},
	})

	finalPath := filepath.Join(rewords, sha+".reword.final.txt")
	if err := os.WriteFile(finalPath, []byte("   \n\n  "), 0o600); err != nil {
		t.Fatalf("write final: %v", err)
	}

	err := rewordApplyMain([]string{"--plan=" + planPath, "--rewords-dir=" + rewords})
	if err == nil {
		t.Fatal("expected error when .reword.final.txt is whitespace-only, got nil")
	}
}
