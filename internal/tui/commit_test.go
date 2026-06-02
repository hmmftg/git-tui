package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"gitflow-tui/internal/git"
)

func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestCommitConfigureAllowsLetterAInMessage(t *testing.T) {
	model := NewCommitModel(git.NewMockGitService(), DefaultStyles())
	model.SetMode(ModeConfigure)

	for _, r := range "add app config" {
		updated, _ := model.Update(runeKey(r))
		model = updated.(*CommitModel)
	}

	if got := model.subjectInput.Value(); got != "add app config" {
		t.Fatalf("expected typed message to be preserved, got %q", got)
	}
	if !model.autoAdd {
		t.Fatal("typing the letter a should not toggle auto-add")
	}
}

func TestCommitConfigureCtrlTTogglesAutoAdd(t *testing.T) {
	model := NewCommitModel(git.NewMockGitService(), DefaultStyles())
	model.SetMode(ModeConfigure)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	model = updated.(*CommitModel)

	if model.autoAdd {
		t.Fatal("expected ctrl+t to toggle auto-add off")
	}
}

func TestCommitMessageCombinesSubjectAndBody(t *testing.T) {
	model := NewCommitModel(git.NewMockGitService(), DefaultStyles())
	model.SetMode(ModeConfigure)
	model.subjectInput.SetValue("subject")
	model.bodyInput.SetValue("body line")

	params := model.GetParameters()
	if got := params["message"]; got != "subject\n\nbody line" {
		t.Fatalf("unexpected commit message: %q", got)
	}
}
