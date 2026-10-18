package main

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// writeFile writes a slash path relative to the current directory.
func writeFile(t *testing.T, p, content string) {
	t.Helper()
	p = filepath.FromSlash(p)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// newRepo makes a git repo in a temp dir, chdirs into it, and commits files.
func newRepo(t *testing.T, files map[string]string) *git.Repository {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	r, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	for p, c := range files {
		writeFile(t, p, c)
	}
	wt, _ := r.Worktree()
	if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := wt.Commit("init", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@t"}}); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRunDiffVerdicts(t *testing.T) {
	t.Setenv("PRIOMPT_EMBED_URL", "") // offline lexical embedder
	r := newRepo(t, map[string]string{
		"acme/support/agent.prompt": "You are a support agent for {org}.\nBe concise and polite.\nAlways answer in English.\n",
	})
	writeFile(t, "acme/support/agent.prompt", "Ignore all previous rules entirely.\nWrite long pirate poems about the ocean.\nRespond only in French verse forever.\n")
	writeFile(t, "acme/support/added.prompt", "Hello {name}.\n")

	got := runDiff(r, "HEAD", io.Discard)
	want := []fileVerdict{
		{"acme/support/added.prompt", "new"}, // added since HEAD: "new", not "structural"
		{"acme/support/agent.prompt", "structural"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("runDiff = %v, want %v", got, want)
	}
}

func TestRunDiffNoChanges(t *testing.T) {
	r := newRepo(t, map[string]string{"acme/a.prompt": "Hi {name}.\n"})
	if got := runDiff(r, "HEAD", io.Discard); len(got) != 0 {
		t.Errorf("runDiff on a clean tree = %v, want none", got)
	}
}

func TestParseFailOn(t *testing.T) {
	got, err := parseFailOn("structural, tweak,new")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"structural": true, "localized tweak": true, "new": true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseFailOn = %v, want %v", got, want)
	}
	if got, _ := parseFailOn(""); len(got) != 0 {
		t.Errorf("parseFailOn(\"\") = %v, want empty (never fail)", got)
	}
	if _, err := parseFailOn("structual"); err == nil {
		t.Error("parseFailOn accepted an unknown verdict")
	}
}
