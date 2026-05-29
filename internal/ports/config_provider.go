package ports

import "github.com/charmbracelet/bubbles/key"

type Keybinding struct {
	Keys   []key.Binding
	Label  string
	Action string
}

type Theme struct {
	Background  string
	Foreground  string
	Accent      string
	Success     string
	Warning     string
	Error       string
	DiffAdded   string
	DiffRemoved string
}

type ConfigProvider interface {
	GetKeybindings() map[string]Keybinding
	GetTheme() Theme
	GetRepoPath() string
	GetMaxLogItems() int
}