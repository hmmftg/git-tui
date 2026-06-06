package workflow

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"gitflow-tui/internal/models"
)

// Store persists and retrieves workflows.
type Store interface {
	Save(models.Workflow) error
	Delete(id string) error
	Get(id string) (*models.Workflow, error)
	List() ([]models.Workflow, error)
}

// ViperStore implements Store using viper-backed YAML config.
type ViperStore struct {
	v    *viper.Viper
	path string
}

// NewViperStore creates a new workflow store backed by viper.
func NewViperStore() *ViperStore {
	v := viper.New()
	v.SetConfigName("gitflow")
	v.SetConfigType("yaml")

	configDir, _ := os.UserConfigDir()
	v.AddConfigPath(filepath.Join(configDir, "gitflow"))
	v.AddConfigPath(".")

	path := filepath.Join(configDir, "gitflow", "gitflow.yaml")

	s := &ViperStore{v: v, path: path}
	_ = s.v.ReadInConfig() // ignore not-found; will be created on first save
	return s
}

// Save persists a workflow.
func (s *ViperStore) Save(wf models.Workflow) error {
	workflows := s.list()
	found := false
	for i, w := range workflows {
		if w.ID == wf.ID {
			workflows[i] = wf
			found = true
			break
		}
	}
	if !found {
		workflows = append(workflows, wf)
	}
	s.v.Set("workflows", workflows)
	if dir := filepath.Dir(s.path); dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}
	return s.v.WriteConfigAs(s.path)
}

// Delete removes a workflow by ID.
func (s *ViperStore) Delete(id string) error {
	workflows := s.list()
	filtered := make([]models.Workflow, 0, len(workflows))
	for _, w := range workflows {
		if w.ID != id {
			filtered = append(filtered, w)
		}
	}
	s.v.Set("workflows", filtered)
	return s.v.WriteConfigAs(s.path)
}

// Get retrieves a workflow by ID.
func (s *ViperStore) Get(id string) (*models.Workflow, error) {
	for _, w := range s.list() {
		if w.ID == id {
			return &w, nil
		}
	}
	return nil, nil
}

// List returns all workflows.
func (s *ViperStore) List() ([]models.Workflow, error) {
	return s.list(), nil
}

func (s *ViperStore) list() []models.Workflow {
	var workflows []models.Workflow
	_ = s.v.UnmarshalKey("workflows", &workflows)
	return workflows
}
