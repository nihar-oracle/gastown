package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetGitState_IgnoresMergedBranchWhenRemoteDefaultRefIsStale(t *testing.T) {
	repoDir := newGitStateTestRepo(t)

	writeTestFile(t, filepath.Join(repoDir, "feature.txt"), "hello from feature\n")
	runGit(t, repoDir, "checkout", "-b", "polecat/merged-cleanup")
	runGit(t, repoDir, "add", "feature.txt")
	runGit(t, repoDir, "commit", "-m", "feature work")

	runGit(t, repoDir, "checkout", "main")
	runGit(t, repoDir, "merge", "--squash", "polecat/merged-cleanup")
	runGit(t, repoDir, "commit", "-m", "squash merge feature locally")
	runGit(t, repoDir, "checkout", "polecat/merged-cleanup")

	state, err := getGitState(repoDir)
	if err != nil {
		t.Fatalf("getGitState() error = %v", err)
	}
	if state.UnpushedCommits != 0 {
		t.Fatalf("UnpushedCommits = %d, want 0", state.UnpushedCommits)
	}
	if !state.Clean {
		t.Fatalf("Clean = false, want true (state=%+v)", *state)
	}
}

func TestGetGitState_ReportsRecoverableUnmergedCommitsWithoutUpstream(t *testing.T) {
	repoDir := newGitStateTestRepo(t)

	writeTestFile(t, filepath.Join(repoDir, "feature.txt"), "still only on feature\n")
	runGit(t, repoDir, "checkout", "-b", "polecat/unmerged")
	runGit(t, repoDir, "add", "feature.txt")
	runGit(t, repoDir, "commit", "-m", "unmerged feature work")

	state, err := getGitState(repoDir)
	if err != nil {
		t.Fatalf("getGitState() error = %v", err)
	}
	if state.UnpushedCommits == 0 {
		t.Fatalf("UnpushedCommits = 0, want > 0 (state=%+v)", *state)
	}
	if state.Clean {
		t.Fatalf("Clean = true, want false (state=%+v)", *state)
	}
}

func newGitStateTestRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	originDir := filepath.Join(root, "origin.git")
	repoDir := filepath.Join(root, "repo")

	runGit(t, root, "init", "--bare", "--initial-branch=main", originDir)
	runGit(t, root, "clone", originDir, repoDir)

	runGit(t, repoDir, "config", "user.name", "Test User")
	runGit(t, repoDir, "config", "user.email", "test@example.com")

	writeTestFile(t, filepath.Join(repoDir, "README.md"), "base\n")
	runGit(t, repoDir, "add", "README.md")
	runGit(t, repoDir, "commit", "-m", "initial commit")
	runGit(t, repoDir, "push", "-u", "origin", "main")

	return repoDir
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
