package git

import (
	"testing"

	"gitflow-tui/internal/models"
)

func TestMockGitService_IsGitRepo(t *testing.T) {
	mock := NewMockGitService()

	if !mock.IsGitRepo() {
		t.Error("Expected IsGitRepo to return true")
	}

	mock.IsRepo = false
	if mock.IsGitRepo() {
		t.Error("Expected IsGitRepo to return false after setting to false")
	}
}

func TestMockGitService_CurrentBranch(t *testing.T) {
	mock := NewMockGitService()

	branch, err := mock.CurrentBranch()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if branch != "main" {
		t.Errorf("Expected branch 'main', got '%s'", branch)
	}
}

func TestMockGitService_GetBranches(t *testing.T) {
	mock := NewMockGitService()

	branches, err := mock.GetBranches()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(branches) != 3 {
		t.Errorf("Expected 3 branches, got %d", len(branches))
	}
}

func TestMockGitService_Status(t *testing.T) {
	mock := NewMockGitService()

	status, err := mock.Status()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if status.Branch != "main" {
		t.Errorf("Expected branch 'main', got '%s'", status.Branch)
	}

	if !status.IsClean {
		t.Error("Expected clean status")
	}
}

func TestMockGitService_Checkout(t *testing.T) {
	mock := NewMockGitService()

	err := mock.Checkout("develop")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	branch, _ := mock.CurrentBranch()
	if branch != "develop" {
		t.Errorf("Expected branch 'develop', got '%s'", branch)
	}
}

func TestMockGitService_Commit(t *testing.T) {
	mock := NewMockGitService()

	err := mock.Commit("test message")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that command was recorded
	if len(mock.Commands) == 0 {
		t.Error("Expected command to be recorded")
	}
}

func TestMockGitService_Push(t *testing.T) {
	mock := NewMockGitService()

	err := mock.Push("origin", "main")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockGitService_Pull(t *testing.T) {
	mock := NewMockGitService()

	err := mock.Pull()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockGitService_Merge(t *testing.T) {
	mock := NewMockGitService()

	err := mock.Merge("develop")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockGitService_Rebase(t *testing.T) {
	mock := NewMockGitService()

	err := mock.Rebase("main")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockGitService_FailureCases(t *testing.T) {
	mock := NewMockGitService()

	// Test commit failure
	mock.ShouldFail["commit"] = true
	err := mock.Commit("test")
	if err == nil {
		t.Error("Expected commit to fail")
	}

	// Reset and test push failure
	mock.ShouldFail = make(map[string]bool)
	mock.ShouldFail["push"] = true
	err = mock.Push("origin", "main")
	if err == nil {
		t.Error("Expected push to fail")
	}
}

func TestStatusStruct(t *testing.T) {
	status := models.Status{
		Branch:   "feature/test",
		Modified: []string{"file1.go", "file2.go"},
		IsClean:  false,
	}

	if status.Branch != "feature/test" {
		t.Errorf("Expected branch 'feature/test', got '%s'", status.Branch)
	}

	if len(status.Modified) != 2 {
		t.Errorf("Expected 2 modified files, got %d", len(status.Modified))
	}

	if status.IsClean {
		t.Error("Expected status to not be clean")
	}
}
