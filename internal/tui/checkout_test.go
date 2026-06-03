package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"gitflow-tui/internal/git"
)

func TestCheckoutConfigureCreateBranchFromSelectedBase(t *testing.T) {
	model := NewCheckoutModel(git.NewMockGitService(), DefaultStyles())
	model.SetMode(ModeConfigure)
	updated, _ := model.Update(branchesLoadedMsg([]string{"main", "develop"}))
	model = updated.(*CheckoutModel)
	model.cursor = 1

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model = updated.(*CheckoutModel)
	if !model.creating || !model.IsInputFocused() {
		t.Fatal("expected create branch input to be focused")
	}
	if model.createBase != "develop" {
		t.Fatalf("expected develop as base, got %q", model.createBase)
	}

	for _, r := range "feature/new" {
		updated, _ = model.Update(runeKey(r))
		model = updated.(*CheckoutModel)
	}
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*CheckoutModel)
	if cmd == nil {
		t.Fatal("expected create branch configuration command")
	}

	msg, ok := cmd().(commandConfiguredMsg)
	if !ok {
		t.Fatalf("expected commandConfiguredMsg, got %T", cmd())
	}
	params := msg.Config.Parameters
	if params["create"] != "true" || params["branch"] != "feature/new" || params["base"] != "develop" {
		t.Fatalf("unexpected checkout create parameters: %#v", params)
	}
}

func TestCheckoutExecuteCreateBranchFromBase(t *testing.T) {
	mock := git.NewMockGitService()
	model := NewCheckoutModel(mock, DefaultStyles())
	model.SetMode(ModeExecute)
	model.SetParameters(map[string]string{"create": "true", "branch": "feature/new", "base": "develop"})

	cmd := model.Execute()
	if cmd == nil {
		t.Fatal("expected create branch execute command")
	}
	if msg := cmd(); msg == nil {
		t.Fatal("expected create branch result message")
	}
	if mock.CurrentBranchVal != "feature/new" {
		t.Fatalf("expected new branch to be current, got %q", mock.CurrentBranchVal)
	}
}
