package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

func TestWorkflowEditorFocusConsistency(t *testing.T) {
	editor := NewWorkflowEditor(git.NewMockGitService(), nil, DefaultStyles(), nil, false)

	assertConsistent := func(label string) {
		expected := editor.nameInput.Focused() || editor.descInput.Focused() || editor.paramInput.Focused()
		if editor.IsInputFocused() != expected {
			t.Fatalf("%s: IsInputFocused mismatch", label)
		}
	}

	assertConsistent("initial")
	updated, _ := editor.Update(tea.KeyMsg{Type: tea.KeyTab})
	editor = updated.(*WorkflowEditorModel)
	assertConsistent("after tab to description")
	updated, _ = editor.Update(tea.KeyMsg{Type: tea.KeyTab})
	editor = updated.(*WorkflowEditorModel)
	assertConsistent("after tab to steps")
	if editor.IsInputFocused() {
		t.Fatal("steps mode should not report focused input")
	}
}

func TestWorkflowEditorParameterEditPrefillsAndPreservesUntilSave(t *testing.T) {
	wf := &models.Workflow{
		Name: "wf",
		Steps: []models.WorkflowStep{{
			Type:       models.StepPush,
			Parameters: map[string]string{"remote": "origin"},
		}},
	}
	editor := NewWorkflowEditor(git.NewMockGitService(), nil, DefaultStyles(), wf, false)
	editor.mode = modeEditStep
	editor.stepCursor = 0
	editor.updateParamKeys()

	updated, _ := editor.Update(tea.KeyMsg{Type: tea.KeyEnter})
	editor = updated.(*WorkflowEditorModel)
	if got := editor.paramInput.Value(); got != "origin" {
		t.Fatalf("expected existing value to be prefilled, got %q", got)
	}

	updated, _ = editor.Update(tea.KeyMsg{Type: tea.KeyEsc})
	editor = updated.(*WorkflowEditorModel)
	if got := editor.workflow.Steps[0].Parameters["remote"]; got != "origin" {
		t.Fatalf("expected cancel to preserve original value, got %q", got)
	}
}

func TestWorkflowEditorCommandEditPreservesExistingParameters(t *testing.T) {
	wf := &models.Workflow{
		Name: "wf",
		Steps: []models.WorkflowStep{{
			Type:       models.StepCommit,
			Parameters: map[string]string{"message": "existing", "autoAdd": "true"},
		}},
	}
	editor := NewWorkflowEditor(git.NewMockGitService(), nil, DefaultStyles(), wf, false)
	editor.mode = modeCommandSelection
	editor.stepCursor = 0
	editor.stepTypeCursor = 0 // commit

	updated, cmd := editor.Update(tea.KeyMsg{Type: tea.KeyEnter})
	editor = updated.(*WorkflowEditorModel)
	if cmd == nil || editor.currentCommandModel == nil {
		t.Fatal("expected command model to be opened")
	}

	commitModel := editor.currentCommandModel.(*CommitModel)
	if got := commitModel.GetParameters()["message"]; got != "existing" {
		t.Fatalf("expected existing message to load, got %q", got)
	}
}
