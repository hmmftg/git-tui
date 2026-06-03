package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

func TestCtrlCQuitsWhileCommitInputFocused(t *testing.T) {
	app := NewApp(git.NewMockGitService())
	app.CurrentView = models.ViewCommit
	app.commitModel = NewCommitModel(app.GitService, app.Styles)
	app.commitModel.subjectInput.Focus()

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected ctrl+c to return a quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg from ctrl+c, got %T", cmd())
	}
}

func TestEscLeavesCommitWhileInputFocused(t *testing.T) {
	app := NewApp(git.NewMockGitService())
	app.CurrentView = models.ViewCommit
	app.commitModel = NewCommitModel(app.GitService, app.Styles)
	app.commitModel.subjectInput.Focus()

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected esc to be handled by commit model")
	}
	if msg := cmd(); msg != viewChangeMsg(models.ViewHome) {
		t.Fatalf("expected esc to navigate home, got %#v", msg)
	}
}
