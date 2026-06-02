package git

import (
	"fmt"
	"os/exec"
	"strings"

	"gitflow-tui/internal/models"
)

// ResolutionChoice represents how to resolve a conflict
type ResolutionChoice int

const (
	Ours ResolutionChoice = iota
	Theirs
	Both
)

// GitService defines the interface for git operations
type GitService interface {
	IsGitRepo() bool
	Status() (models.Status, error)
	Commit(message string) error
	Push(remote, branch string) error
	Pull() error
	Checkout(branch string) error
	Merge(branch string) error
	Rebase(branch string) error
	CurrentBranch() (string, error)
	GetBranches() ([]string, error)
	Add(files []string) error
	AddAll() error
	GetLog(limit int) ([]string, error)
	GetGraph() ([]string, error)
	HasConflicts() bool
	AbortMerge() error
	ContinueRebase() error
	AbortRebase() error

	// Conflict resolution methods
	GetConflictedFiles() ([]string, error)
	ReadFile(path string) (string, error)
	ResolveConflict(path string, choice ResolutionChoice) error
}

// service implements GitService
type service struct {
	repoPath string
}

// NewGitService creates a new git service
func NewGitService(path string) GitService {
	return &service{repoPath: path}
}

// exec runs a git command and returns the output
func (s *service) exec(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.repoPath
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// IsGitRepo checks if the current directory is a git repository
func (s *service) IsGitRepo() bool {
	output, err := s.exec("rev-parse", "--git-dir")
	return err == nil && strings.TrimSpace(output) != ""
}

// Status returns the current git status
func (s *service) Status() (models.Status, error) {
	var status models.Status

	// Get current branch
	branch, err := s.CurrentBranch()
	if err != nil {
		return status, err
	}
	status.Branch = branch

	// Check for merge/rebase in progress
	s.checkState(&status)

	// Parse porcelain status
	output, err := s.exec("status", "--porcelain", "-b")
	if err != nil {
		return status, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse branch info from first line
		if strings.HasPrefix(line, "##") {
			status.Branch, status.Ahead, status.Behind = s.parseBranchLine(line)
			continue
		}

		// Parse file status
		if len(line) >= 2 {
			x := line[0] // staged status
			y := line[1] // unstaged status
			file := line[3:]

			if x == 'M' || y == 'M' {
				status.Modified = append(status.Modified, file)
			}
			if x == 'A' {
				status.Added = append(status.Added, file)
			}
			if x == 'D' || y == 'D' {
				status.Deleted = append(status.Deleted, file)
			}
			if x == '?' {
				status.Untracked = append(status.Untracked, file)
			}
			if x == 'R' {
				status.Renamed = append(status.Renamed, file)
			}
			if x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
				status.Conflicted = append(status.Conflicted, file)
			}
		}
	}

	status.IsClean = len(status.Modified) == 0 && len(status.Added) == 0 &&
		len(status.Deleted) == 0 && len(status.Untracked) == 0 &&
		len(status.Renamed) == 0 && len(status.Conflicted) == 0

	return status, nil
}

func (s *service) parseBranchLine(line string) (branch string, ahead, behind int) {
	line = strings.TrimPrefix(line, "## ")

	// Handle detached HEAD
	if strings.HasPrefix(line, "HEAD (no branch)") {
		return "(detached HEAD)", 0, 0
	}

	// Parse branch name and tracking info
	if idx := strings.Index(line, "..."); idx != -1 {
		branch = line[:idx]
		rest := line[idx:]

		// Parse ahead/behind
		if start := strings.Index(rest, "[ahead "); start != -1 {
			end := strings.Index(rest[start:], "]")
			if end != -1 {
				fmt.Sscanf(rest[start:start+end], "[ahead %d", &ahead)
			}
		}
		if start := strings.Index(rest, "[behind "); start != -1 {
			end := strings.Index(rest[start:], "]")
			if end != -1 {
				fmt.Sscanf(rest[start:start+end], "[behind %d", &behind)
			}
		}
	} else {
		branch = strings.Fields(line)[0]
	}

	return branch, ahead, behind
}

func (s *service) checkState(status *models.Status) {
	// Check for merge in progress
	_, err := s.exec("merge", "HEAD")
	status.IsMergeInProgress = err == nil

	// Check for rebase in progress
	output, _ := s.exec("rev-parse", "--git-path", "rebase-merge")
	status.IsRebaseInProgress = output != ""
}

// Commit creates a commit with the given message
func (s *service) Commit(message string) error {
	_, err := s.exec("commit", "-m", message)
	return err
}

// Push pushes changes to remote
func (s *service) Push(remote, branch string) error {
	args := []string{"push"}
	if remote != "" {
		args = append(args, remote)
		if branch != "" {
			args = append(args, branch)
		}
	}
	_, err := s.exec(args...)
	return err
}

// Pull pulls changes from remote
func (s *service) Pull() error {
	_, err := s.exec("pull")
	return err
}

// Checkout switches to the given branch
func (s *service) Checkout(branch string) error {
	_, err := s.exec("checkout", branch)
	return err
}

// Merge merges the given branch into current
func (s *service) Merge(branch string) error {
	_, err := s.exec("merge", branch)
	return err
}

// Rebase rebases current branch onto given branch
func (s *service) Rebase(branch string) error {
	_, err := s.exec("rebase", branch)
	return err
}

// CurrentBranch returns the current branch name
func (s *service) CurrentBranch() (string, error) {
	output, err := s.exec("branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// GetBranches returns all local branches
func (s *service) GetBranches() ([]string, error) {
	output, err := s.exec("branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	branches := strings.Split(strings.TrimSpace(output), "\n")
	return branches, nil
}

// Add stages specific files
func (s *service) Add(files []string) error {
	args := append([]string{"add"}, files...)
	_, err := s.exec(args...)
	return err
}

// AddAll stages all changes
func (s *service) AddAll() error {
	_, err := s.exec("add", ".")
	return err
}

// GetLog returns recent commit history
func (s *service) GetLog(limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}
	output, err := s.exec("log", fmt.Sprintf("-%d", limit), "--oneline", "--decorate")
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(output), "\n"), nil
}

// GetGraph returns a graphical representation of the commit history
func (s *service) GetGraph() ([]string, error) {
	output, err := s.exec("log", "--graph", "--oneline", "--decorate", "--all", "-20")
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(output), "\n"), nil
}

// HasConflicts checks if there are merge conflicts
func (s *service) HasConflicts() bool {
	output, err := s.exec("diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) != ""
}

// AbortMerge aborts the current merge
func (s *service) AbortMerge() error {
	_, err := s.exec("merge", "--abort")
	return err
}

// ContinueRebase continues the current rebase
func (s *service) ContinueRebase() error {
	_, err := s.exec("rebase", "--continue")
	return err
}

// AbortRebase aborts the current rebase
func (s *service) AbortRebase() error {
	_, err := s.exec("rebase", "--abort")
	return err
}

// GetConflictedFiles returns list of files with merge conflicts
func (s *service) GetConflictedFiles() ([]string, error) {
	output, err := s.exec("diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	files := strings.Split(strings.TrimSpace(output), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}
	return files, nil
}

// ReadFile reads the content of a file
func (s *service) ReadFile(path string) (string, error) {
	output, err := s.exec("show", fmt.Sprintf(":/%s", path))
	if err != nil {
		// Try reading from working directory if not in index
		content, err := s.exec("cat-file", "-p", fmt.Sprintf("HEAD:%s", path))
		if err != nil {
			// Last resort - read directly from filesystem
			cmd := exec.Command("cmd", "/c", "type", path)
			cmd.Dir = s.repoPath
			out, err := cmd.CombinedOutput()
			if err != nil {
				return "", err
			}
			return string(out), nil
		}
		return content, nil
	}
	return output, nil
}

// ResolveConflict resolves a conflicted file using the specified choice
func (s *service) ResolveConflict(path string, choice ResolutionChoice) error {
	switch choice {
	case Ours:
		_, err := s.exec("checkout", "--ours", path)
		if err != nil {
			return err
		}
	case Theirs:
		_, err := s.exec("checkout", "--theirs", path)
		if err != nil {
			return err
		}
	case Both:
		// Keep both (conflict markers) - just stage as-is
		// No checkout needed
	}

	// Stage the resolved file
	_, err := s.exec("add", path)
	return err
}
