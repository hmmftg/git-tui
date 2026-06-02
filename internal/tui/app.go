package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"gitflow-tui/internal/config"
	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// AppModel is the main application model
type AppModel struct {
	CurrentView models.ViewType
	Repository  models.RepositoryInfo
	GitService  git.GitService
	ConfigMgr   *config.Manager
	Styles      Styles
	Width       int
	Height      int

	// View models
	statusModel           *StatusModel
	commitModel           *CommitModel
	checkoutModel         *CheckoutModel
	pushModel             *PushModel
	pullModel             *PullModel
	mergeModel            *MergeModel
	rebaseModel           *RebaseModel
	workflowModel         *WorkflowModel
	workflowEditorModel   *WorkflowEditorModel
	conflictResolverModel *ConflictResolverModel
	executionModel        *ExecutionModel

	// Message/Error
	Message     string
	MessageType string // "error", "success", "info"
}

// NewApp creates a new TUI application
func NewApp(gitSvc git.GitService) *AppModel {
	// Initialize config manager
	cfg := config.NewManager()
	cfg.Load() // Try to load existing config

	return &AppModel{
		CurrentView: models.ViewHome,
		GitService:  gitSvc,
		ConfigMgr:   cfg,
		Styles:      DefaultStyles(),
	}
}

// Init initializes the application
func (m *AppModel) Init() tea.Cmd {
	// Initialize repository info
	return m.refreshRepoInfo()
}

// Update handles messages and updates the model
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.updateChildSizes()
		return m, nil

	case tea.KeyMsg:
		inputFocused := m.isInputFocused()

		// Ctrl+C is always global so users can reliably exit focused inputs.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Let focused inputs and form models consume their local editing keys.
		if inputFocused {
			break
		}

		// Global navigation keys (only when no input is focused)
		switch msg.String() {
		case "q":
			if m.CurrentView == models.ViewHome {
				return m, tea.Quit
			}
			return m, m.navigateTo(models.ViewHome)
		case "esc":
			if m.CurrentView != models.ViewHome {
				return m, m.navigateTo(models.ViewHome)
			}
			return m, tea.Quit
		case "home":
			return m, m.navigateTo(models.ViewHome)
		case "f1", "?":
			// Show help
			return m, nil
		// Numeric keys for home menu navigation
		case "1":
			if m.CurrentView == models.ViewHome {
				return m, m.navigateTo(models.ViewStatus)
			}
		case "2":
			if m.CurrentView == models.ViewHome {
				return m, m.navigateTo(models.ViewCommit)
			}
		case "3":
			if m.CurrentView == models.ViewHome {
				return m, m.navigateTo(models.ViewPush)
			}
		case "4":
			if m.CurrentView == models.ViewHome {
				return m, m.navigateTo(models.ViewPull)
			}
		case "5":
			if m.CurrentView == models.ViewHome {
				return m, m.navigateTo(models.ViewCheckout)
			}
		case "6":
			if m.CurrentView == models.ViewHome {
				return m, m.navigateTo(models.ViewMerge)
			}
		case "7":
			if m.CurrentView == models.ViewHome {
				return m, m.navigateTo(models.ViewRebase)
			}
		case "8":
			if m.CurrentView == models.ViewHome {
				return m, m.navigateTo(models.ViewWorkflowBuilder)
			}
		}

	case repoInfoMsg:
		m.Repository = models.RepositoryInfo(msg)
		return m, nil

	case runWorkflowMsg:
		// Start workflow execution
		m.executionModel = NewExecutionModel(m.GitService, m.Styles, models.Workflow(msg), models.StopOnError)
		return m, tea.Batch(
			m.navigateTo(models.ViewExecution),
			m.executionModel.Init(),
		)

	case createWorkflowMsg:
		// Open workflow editor for new workflow (from template or blank)
		var wf *models.Workflow
		if msg.fromTemplate != nil {
			copy := *msg.fromTemplate
			copy.ID = uuid.New().String()
			copy.Name = copy.Name + " Copy"
			copy.CreatedAt = time.Now()
			wf = &copy
		}
		m.workflowEditorModel = NewWorkflowEditor(m.GitService, m.ConfigMgr, m.Styles, wf, wf != nil)
		m.updateChildSizes()
		return m, tea.Batch(
			m.navigateTo(models.ViewWorkflowEditor),
			m.workflowEditorModel.Init(),
		)

	case editWorkflowMsg:
		// Open workflow editor for existing workflow
		wf := models.Workflow(msg)
		m.workflowEditorModel = NewWorkflowEditor(m.GitService, m.ConfigMgr, m.Styles, &wf, false)
		m.updateChildSizes()
		return m, tea.Batch(
			m.navigateTo(models.ViewWorkflowEditor),
			m.workflowEditorModel.Init(),
		)

	case saveWorkflowMsg:
		// Workflow saved, return to workflow list and refresh
		m.Message = fmt.Sprintf("Workflow saved: %s", models.Workflow(msg).Name)
		m.MessageType = "success"
		m.workflowEditorModel = nil
		return m, m.navigateTo(models.ViewWorkflowBuilder)

	case cancelEditMsg:
		// Cancel editing, return to workflow list
		m.workflowEditorModel = nil
		return m, m.navigateTo(models.ViewWorkflowBuilder)

	case openConflictResolverMsg:
		// Open conflict resolver
		m.conflictResolverModel = NewConflictResolver(m.GitService, m.Styles)
		return m, tea.Batch(
			m.navigateTo(models.ViewConflictResolver),
			m.conflictResolverModel.Init(),
		)

	case conflictsResolvedMsg:
		// Conflicts resolved, return to merge/rebase
		m.conflictResolverModel = nil
		m.Message = "✔ All conflicts resolved"
		m.MessageType = "success"
		// Try to continue the operation
		if m.CurrentView == models.ViewMerge || m.CurrentView == models.ViewConflictResolver {
			return m, tea.Batch(
				m.navigateTo(models.ViewMerge),
				func() tea.Msg { return MergeContinueMsg{} },
			)
		} else if m.CurrentView == models.ViewRebase {
			return m, tea.Batch(
				m.navigateTo(models.ViewRebase),
				func() tea.Msg { return RebaseContinueMsg{} },
			)
		}
		return m, nil

	case conflictsAbortedMsg:
		// Conflicts resolution aborted
		m.conflictResolverModel = nil
		m.Message = "✔ Conflict resolution aborted"
		m.MessageType = "info"
		if m.CurrentView == models.ViewMerge {
			return m, tea.Batch(
				m.navigateTo(models.ViewMerge),
				func() tea.Msg { return MergeAbortedMsg{} },
			)
		} else if m.CurrentView == models.ViewRebase {
			return m, tea.Batch(
				m.navigateTo(models.ViewRebase),
				func() tea.Msg { return RebaseAbortedMsg{} },
			)
		}
		return m, nil

	case conflictsCancelledMsg:
		// User cancelled resolver without resolving
		m.conflictResolverModel = nil
		if m.CurrentView == models.ViewMerge {
			return m, m.navigateTo(models.ViewMerge)
		}
		return m, m.navigateTo(models.ViewRebase)

	case viewChangeMsg:
		m.CurrentView = models.ViewType(msg)
		// Initialize view-specific models
		switch m.CurrentView {
		case models.ViewStatus:
			m.statusModel = NewStatusModel(m.GitService, m.Styles)
			return m, m.statusModel.Init()
		case models.ViewCommit:
			m.commitModel = NewCommitModel(m.GitService, m.Styles)
			m.updateChildSizes()
			return m, m.commitModel.Init()
		case models.ViewCheckout:
			m.checkoutModel = NewCheckoutModel(m.GitService, m.Styles)
			return m, m.checkoutModel.Init()
		case models.ViewPush:
			m.pushModel = NewPushModel(m.GitService, m.Styles)
			return m, m.pushModel.Init()
		case models.ViewPull:
			m.pullModel = NewPullModel(m.GitService, m.Styles)
			return m, m.pullModel.Init()
		case models.ViewMerge:
			m.mergeModel = NewMergeModel(m.GitService, m.Styles)
			return m, m.mergeModel.Init()
		case models.ViewRebase:
			m.rebaseModel = NewRebaseModel(m.GitService, m.Styles)
			return m, m.rebaseModel.Init()
		case models.ViewConflictResolver:
			// Conflict resolver is already set by openConflictResolverMsg handler
			return m, nil
		case models.ViewWorkflowBuilder:
			m.workflowModel = NewWorkflowModel(m.GitService, m.ConfigMgr, m.Styles)
			return m, m.workflowModel.Init()
		case models.ViewWorkflowEditor:
			// Workflow editor model is already set by createWorkflowMsg/editWorkflowMsg handlers
			return m, nil
		case models.ViewExecution:
			// Execution model is already set by runWorkflowMsg handler
			return m, nil
		}
		return m, nil

	case errorMsg:
		m.Message = string(msg)
		m.MessageType = "error"
		return m, nil

	case successMsg:
		m.Message = string(msg)
		m.MessageType = "success"
		return m, nil
	}

	// Route to view-specific models
	var cmd tea.Cmd
	switch m.CurrentView {
	case models.ViewStatus:
		if m.statusModel != nil {
			sm, c := m.statusModel.Update(msg)
			m.statusModel = sm.(*StatusModel)
			cmd = c
		}
	case models.ViewCommit:
		if m.commitModel != nil {
			cm, c := m.commitModel.Update(msg)
			m.commitModel = cm.(*CommitModel)
			cmd = c
		}
	case models.ViewCheckout:
		if m.checkoutModel != nil {
			cm, c := m.checkoutModel.Update(msg)
			m.checkoutModel = cm.(*CheckoutModel)
			cmd = c
		}
	case models.ViewPush:
		if m.pushModel != nil {
			pm, c := m.pushModel.Update(msg)
			m.pushModel = pm.(*PushModel)
			cmd = c
		}
	case models.ViewPull:
		if m.pullModel != nil {
			pm, c := m.pullModel.Update(msg)
			m.pullModel = pm.(*PullModel)
			cmd = c
		}
	case models.ViewMerge:
		if m.mergeModel != nil {
			mm, c := m.mergeModel.Update(msg)
			m.mergeModel = mm.(*MergeModel)
			cmd = c
		}
	case models.ViewRebase:
		if m.rebaseModel != nil {
			rm, c := m.rebaseModel.Update(msg)
			m.rebaseModel = rm.(*RebaseModel)
			cmd = c
		}
	case models.ViewConflictResolver:
		if m.conflictResolverModel != nil {
			crm, c := m.conflictResolverModel.Update(msg)
			m.conflictResolverModel = crm.(*ConflictResolverModel)
			cmd = c
		}
	case models.ViewWorkflowBuilder:
		if m.workflowModel != nil {
			wm, c := m.workflowModel.Update(msg)
			m.workflowModel = wm.(*WorkflowModel)
			cmd = c
		}
	case models.ViewWorkflowEditor:
		if m.workflowEditorModel != nil {
			em, c := m.workflowEditorModel.Update(msg)
			m.workflowEditorModel = em.(*WorkflowEditorModel)
			cmd = c
		}
	case models.ViewExecution:
		if m.executionModel != nil {
			em, c := m.executionModel.Update(msg)
			m.executionModel = em.(*ExecutionModel)
			cmd = c
		}
	}

	return m, cmd
}

// View renders the current view
func (m *AppModel) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	var content string

	switch m.CurrentView {
	case models.ViewHome:
		content = m.renderHome()
	case models.ViewStatus:
		if m.statusModel != nil {
			content = m.statusModel.View()
		}
	case models.ViewCommit:
		if m.commitModel != nil {
			content = m.commitModel.View()
		}
	case models.ViewCheckout:
		if m.checkoutModel != nil {
			content = m.checkoutModel.View()
		}
	case models.ViewPush:
		if m.pushModel != nil {
			content = m.pushModel.View()
		}
	case models.ViewPull:
		if m.pullModel != nil {
			content = m.pullModel.View()
		}
	case models.ViewMerge:
		if m.mergeModel != nil {
			content = m.mergeModel.View()
		}
	case models.ViewRebase:
		if m.rebaseModel != nil {
			content = m.rebaseModel.View()
		}
	case models.ViewConflictResolver:
		if m.conflictResolverModel != nil {
			content = m.conflictResolverModel.View()
		}
	case models.ViewWorkflowBuilder:
		if m.workflowModel != nil {
			content = m.workflowModel.View()
		}
	case models.ViewWorkflowEditor:
		if m.workflowEditorModel != nil {
			content = m.workflowEditorModel.View()
		}
	case models.ViewExecution:
		if m.executionModel != nil {
			content = m.executionModel.View()
		}
	default:
		content = m.renderHome()
	}

	// Build the full layout
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		content,
		m.renderFooter(),
	)
}

// renderHeader renders the application header
func (m *AppModel) renderHeader() string {
	repoInfo := "Not in a repository"
	if m.Repository.IsValid {
		repoInfo = fmt.Sprintf("📁 %s | 🔀 %s", m.Repository.Path, m.Repository.Status.Branch)
	}

	headerStyle := m.Styles.Header
	if m.Width > 0 {
		headerStyle = headerStyle.Width(m.Width)
	}
	return headerStyle.Render(fmt.Sprintf(" GitFlow TUI | %s ", repoInfo))
}

// renderFooter renders the application footer
func (m *AppModel) renderFooter() string {
	help := "q: quit | esc: home | ? help"
	if m.Message != "" {
		var msgStyle lipgloss.Style
		switch m.MessageType {
		case "error":
			msgStyle = m.Styles.Error
		case "success":
			msgStyle = m.Styles.Success
		default:
			msgStyle = m.Styles.Info
		}
		help = fmt.Sprintf("%s | %s", msgStyle.Render(m.Message), help)
	}

	footerStyle := m.Styles.Footer
	if m.Width > 0 {
		footerStyle = footerStyle.Width(m.Width)
	}
	return footerStyle.Render(help)
}

// renderHome renders the home view
func (m *AppModel) renderHome() string {
	menuItems := []struct {
		key   string
		label string
		view  models.ViewType
	}{
		{"1", "📊 Status", models.ViewStatus},
		{"2", "📝 Commit", models.ViewCommit},
		{"3", "📤 Push", models.ViewPush},
		{"4", "📥 Pull", models.ViewPull},
		{"5", "🔀 Checkout", models.ViewCheckout},
		{"6", "⛙ Merge", models.ViewMerge},
		{"7", "🔄 Rebase", models.ViewRebase},
		{"8", "⚡ Workflows", models.ViewWorkflowBuilder},
		{"q", "❌ Quit", models.ViewHome},
	}

	var menu []string
	menu = append(menu, "")

	for _, item := range menuItems {
		line := fmt.Sprintf(" [%s] %s", m.Styles.Key.Render(item.key), item.label)
		menu = append(menu, m.Styles.MenuItem.Render(line))
	}

	return lipgloss.Place(
		m.Width,
		m.Height-lipgloss.Height(m.renderHeader())-lipgloss.Height(m.renderFooter()),
		lipgloss.Center,
		lipgloss.Center,
		m.Styles.Menu.Render(lipgloss.JoinVertical(lipgloss.Left, menu...)),
	)
}

// navigateTo changes the current view
func (m *AppModel) navigateTo(view models.ViewType) tea.Cmd {
	return func() tea.Msg {
		return viewChangeMsg(view)
	}
}

// refreshRepoInfo refreshes repository information
func (m *AppModel) refreshRepoInfo() tea.Cmd {
	return func() tea.Msg {
		if !m.GitService.IsGitRepo() {
			return repoInfoMsg(models.RepositoryInfo{IsValid: false})
		}

		status, err := m.GitService.Status()
		if err != nil {
			return errorMsg(err.Error())
		}

		// Get repo path
		cmd := exec.Command("git", "rev-parse", "--show-toplevel")
		path, _ := cmd.Output()

		return repoInfoMsg(models.RepositoryInfo{
			Path:    strings.TrimSpace(string(path)),
			IsValid: true,
			Status:  status,
		})
	}
}

func (m *AppModel) updateChildSizes() {
	contentWidth := m.Width - 6
	contentHeight := m.Height - lipgloss.Height(m.renderHeader()) - lipgloss.Height(m.renderFooter())
	if m.commitModel != nil {
		m.commitModel.SetSize(contentWidth, contentHeight)
	}
	if m.workflowEditorModel != nil {
		m.workflowEditorModel.SetSize(contentWidth, contentHeight)
	}
}

// isInputFocused checks if any text input is currently focused in the current view
func (m *AppModel) isInputFocused() bool {
	switch m.CurrentView {
	case models.ViewWorkflowEditor:
		if m.workflowEditorModel != nil {
			return m.workflowEditorModel.IsInputFocused()
		}
	case models.ViewCommit:
		if m.commitModel != nil {
			return m.commitModel.IsInputFocused()
		}
	}
	return false
}

// Message types
type repoInfoMsg models.RepositoryInfo
type viewChangeMsg models.ViewType
type errorMsg string
type successMsg string
type openConflictResolverMsg struct{}
