package fileconfig

import (
	"log"
	"os"

	"github.com/sazardev/gitto/internal/ports"
)

type Adapter struct {
	config *Config
}

type Config struct {
	RepoPath    string
	MaxLogItems int
	Keybindings map[string]KeybindingConfig
	Theme       ThemeConfig
}

type KeybindingConfig struct {
	Keys   []string
	Label  string
	Action string
}

type ThemeConfig struct {
	Background   string
	Foreground   string
	Accent       string
	Success      string
	Warning      string
	Error        string
	DiffAdded    string
	DiffRemoved  string
}

func NewAdapter(configPath string) (*Adapter, error) {
	_, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Warning: config file not found at %s, using defaults", configPath)
	}

	return &Adapter{config: &Config{
		RepoPath:    ".",
		MaxLogItems: 50,
		Keybindings: make(map[string]KeybindingConfig),
		Theme: ThemeConfig{
			Background:  "#1a1a2e",
			Foreground:  "#eaeaea",
			Accent:      "#7c3aed",
			Success:     "#22c55e",
			Warning:     "#f59e0b",
			Error:       "#ef4444",
			DiffAdded:   "#22c55e",
			DiffRemoved: "#ef4444",
		},
	}}, nil
}

func (a *Adapter) GetKeybindings() map[string]ports.Keybinding {
	result := make(map[string]ports.Keybinding)
	for name, kb := range a.config.Keybindings {
		result[name] = ports.Keybinding{
			Label:  kb.Label,
			Action: kb.Action,
		}
	}
	return result
}

func (a *Adapter) GetTheme() ports.Theme {
	return ports.Theme{
		Background:  a.config.Theme.Background,
		Foreground:  a.config.Theme.Foreground,
		Accent:      a.config.Theme.Accent,
		Success:     a.config.Theme.Success,
		Warning:     a.config.Theme.Warning,
		Error:       a.config.Theme.Error,
		DiffAdded:   a.config.Theme.DiffAdded,
		DiffRemoved: a.config.Theme.DiffRemoved,
	}
}

func (a *Adapter) GetRepoPath() string {
	return a.config.RepoPath
}

func (a *Adapter) GetMaxLogItems() int {
	return a.config.MaxLogItems
}

var _ ports.ConfigProvider = (*Adapter)(nil)