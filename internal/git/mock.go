package git

import (
	"errors"
	"fmt"
	"gitflow-tui/internal/models"
)

// MockGitService is a mock implementation for testing
type MockGitService struct {
	IsRepo           bool
	CurrentBranchVal string
	StatusVal        models.Status
	StatusErr        error
	Branches         []string
	Commands         []string
	ShouldFail       map[string]bool
}

// NewMockGitService creates a new mock service
func NewMockGitService() *MockGitService {
	return &MockGitService{
		IsRepo:           true,
		CurrentBranchVal: "main",
		StatusVal: models.Status{
			Branch:    "main",
			IsClean:   true,
			Modified:  []string{},
			Added:     []string{},
			Deleted:   []string{},
			Untracked: []string{},
		},
		Branches:   []string{"main", "develop", "feature/test"},
		Commands:   []string{},
		ShouldFail: make(map[string]bool),
	}
}

func (m *MockGitService) record(args ...string) {
	m.Commands = append(m.Commands, fmt.Sprintf("git %v", args))
}

func (m *MockGitService) IsGitRepo() bool {
	return m.IsRepo
}

func (m *MockGitService) Status() (models.Status, error) {
	m.record("status")
	if m.StatusErr != nil {
		return models.Status{}, m.StatusErr
	}
	return m.StatusVal, nil
}

func (m *MockGitService) Commit(message string) error {
	m.record("commit", "-m", message)
	if m.ShouldFail["commit"] {
		return errors.New("commit failed")
	}
	return nil
}

func (m *MockGitService) Push(remote, branch string) error {
	m.record("push", remote, branch)
	if m.ShouldFail["push"] {
		return errors.New("push failed")
	}
	return nil
}

func (m *MockGitService) Pull() error {
	m.record("pull")
	if m.ShouldFail["pull"] {
		return errors.New("pull failed")
	}
	return nil
}

func (m *MockGitService) Checkout(branch string) error {
	m.record("checkout", branch)
	if m.ShouldFail["checkout"] {
		return errors.New("checkout failed")
	}
	m.CurrentBranchVal = branch
	m.StatusVal.Branch = branch
	return nil
}

func (m *MockGitService) Merge(branch string) error {
	m.record("merge", branch)
	if m.ShouldFail["merge"] {
		return errors.New("merge failed")
	}
	return nil
}

func (m *MockGitService) Rebase(branch string) error {
	m.record("rebase", branch)
	if m.ShouldFail["rebase"] {
		return errors.New("rebase failed")
	}
	return nil
}

func (m *MockGitService) CurrentBranch() (string, error) {
	m.record("branch", "--show-current")
	if m.ShouldFail["current"] {
		return "", errors.New("failed to get branch")
	}
	return m.CurrentBranchVal, nil
}

func (m *MockGitService) GetBranches() ([]string, error) {
	m.record("branch")
	if m.ShouldFail["branches"] {
		return nil, errors.New("failed to get branches")
	}
	return m.Branches, nil
}

func (m *MockGitService) Add(files []string) error {
	m.record(append([]string{"add"}, files...)...)
	if m.ShouldFail["add"] {
		return errors.New("add failed")
	}
	return nil
}

func (m *MockGitService) AddAll() error {
	m.record("add", ".")
	if m.ShouldFail["add"] {
		return errors.New("add failed")
	}
	return nil
}

func (m *MockGitService) GetLog(limit int) ([]string, error) {
	m.record("log")
	if m.ShouldFail["log"] {
		return nil, errors.New("log failed")
	}
	return []string{
		"abc1234 Initial commit",
		"def5678 Add feature X",
		"ghi9012 Fix bug Y",
	}, nil
}

func (m *MockGitService) GetGraph() ([]string, error) {
	m.record("log", "--graph")
	if m.ShouldFail["graph"] {
		return nil, errors.New("graph failed")
	}
	return []string{
		"* abc1234 Initial commit",
		"* def5678 Add feature X",
		"* ghi9012 Fix bug Y",
	}, nil
}

func (m *MockGitService) HasConflicts() bool {
	m.record("diff", "--name-only", "--diff-filter=U")
	return len(m.StatusVal.Conflicted) > 0
}

func (m *MockGitService) AbortMerge() error {
	m.record("merge", "--abort")
	if m.ShouldFail["abort-merge"] {
		return errors.New("abort merge failed")
	}
	return nil
}

func (m *MockGitService) ContinueRebase() error {
	m.record("rebase", "--continue")
	if m.ShouldFail["continue-rebase"] {
		return errors.New("continue rebase failed")
	}
	return nil
}

func (m *MockGitService) AbortRebase() error {
	m.record("rebase", "--abort")
	if m.ShouldFail["abort-rebase"] {
		return errors.New("abort rebase failed")
	}
	return nil
}

// Verify that MockGitService implements GitService
var _ GitService = (*MockGitService)(nil)
