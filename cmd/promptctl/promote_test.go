package main

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestPromoteRestoresBranch(t *testing.T) {
	const devContent = "Hello {name}, welcome to {org}.\n"
	dir := t.TempDir()
	t.Chdir(dir)
	r, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	wt, _ := r.Worktree()
	commit := func(content, msg string) {
		t.Helper()
		if err := os.MkdirAll("acme", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("acme/a.prompt", []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := wt.Add("acme/a.prompt"); err != nil {
			t.Fatal(err)
		}
		if _, err := wt.Commit(msg, &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@t"}}); err != nil {
			t.Fatal(err)
		}
	}
	commit("Hello {name}.\n", "init")
	head, _ := r.Head()
	for _, b := range []string{"staging", "dev"} {
		if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName(b), head.Hash())); err != nil {
			t.Fatal(err)
		}
	}
	if err := wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("dev")}); err != nil {
		t.Fatal(err)
	}
	commit(devContent, "dev edit")

	promote([]string{"acme/a.prompt", "dev", "staging"})

	if h, _ := r.Head(); h.Name().Short() != "dev" {
		t.Errorf("after promote HEAD is %s, want dev (the branch it started on)", h.Name().Short())
	}
	if got := diskContent("acme/a.prompt"); got != devContent {
		t.Errorf("working tree has %q, want dev's content", got)
	}
	if got := blobContent(treeAt(r, "staging"), "acme/a.prompt"); got != devContent {
		t.Errorf("staging has %q, want the promoted content", got)
	}
}
