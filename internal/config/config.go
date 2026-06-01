package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"gitflow-tui/internal/models"
)

// Manager handles configuration
type Manager struct {
	viper *viper.Viper
	path  string
}

// NewManager creates a new config manager
func NewManager() *Manager {
	v := viper.New()
	v.SetConfigName("gitflow")
	v.SetConfigType("yaml")

	// Set default paths
	configDir, _ := os.UserConfigDir()
	v.AddConfigPath(filepath.Join(configDir, "gitflow"))
	v.AddConfigPath(".")

	return &Manager{
		viper: v,
		path:  filepath.Join(configDir, "gitflow", "gitflow.yaml"),
	}
}

// Load loads configuration from file
func (m *Manager) Load() error {
	// Ensure config directory exists
	if dir := filepath.Dir(m.path); dir != "" {
		os.MkdirAll(dir, 0755)
	}

	if err := m.viper.ReadInConfig(); err != nil {
		// Config file not found is OK, we'll create it
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return nil
}

// Save saves configuration to file
func (m *Manager) Save() error {
	return m.viper.WriteConfigAs(m.path)
}

// GetWorkflows returns all saved workflows
func (m *Manager) GetWorkflows() []models.Workflow {
	var workflows []models.Workflow
	if err := m.viper.UnmarshalKey("workflows", &workflows); err != nil {
		return nil
	}
	return workflows
}

// SaveWorkflow saves a workflow to config
func (m *Manager) SaveWorkflow(workflow models.Workflow) error {
	workflows := m.GetWorkflows()
	
	// Check if workflow already exists
	found := false
	for i, w := range workflows {
		if w.ID == workflow.ID {
			workflows[i] = workflow
			found = true
			break
		}
	}

	if !found {
		workflows = append(workflows, workflow)
	}

	m.viper.Set("workflows", workflows)
	return m.Save()
}

// DeleteWorkflow removes a workflow from config
func (m *Manager) DeleteWorkflow(id string) error {
	workflows := m.GetWorkflows()
	
	newWorkflows := make([]models.Workflow, 0, len(workflows))
	for _, w := range workflows {
		if w.ID != id {
			newWorkflows = append(newWorkflows, w)
		}
	}

	m.viper.Set("workflows", newWorkflows)
	return m.Save()
}

// GetSetting gets a setting value
func (m *Manager) GetSetting(key string) interface{} {
	return m.viper.Get(key)
}

// SetSetting sets a setting value
func (m *Manager) SetSetting(key string, value interface{}) {
	m.viper.Set(key, value)
}

// DefaultConfig returns the default configuration
func DefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"workflows": []models.Workflow{
			{
				Name:        "Quick Commit",
				Description: "Stage, commit and push all changes",
				Steps: []models.WorkflowStep{
					{
						Type:        models.StepStatus,
						Description: "Check repository status",
					},
					{
						Type:        models.StepCommit,
						Description: "Commit all changes",
						Parameters: map[string]string{
							"autoAdd": "true",
							"message":  "WIP: auto commit",
						},
					},
					{
						Type:        models.StepPush,
						Description: "Push to remote",
					},
				},
			},
			{
				Name:        "Sync",
				Description: "Pull latest changes from remote",
				Steps: []models.WorkflowStep{
					{
						Type:        models.StepStatus,
						Description: "Check current status",
					},
					{
						Type:        models.StepPull,
						Description: "Pull from remote",
					},
				},
			},
		},
	}
}
