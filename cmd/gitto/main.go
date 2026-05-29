package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/sazardev/gitto/internal/adapters/fileconfig"
	"github.com/sazardev/gitto/internal/adapters/gogit"
	"github.com/sazardev/gitto/internal/ui"
)

func main() {
	repoPath := getRepoPath()
	configPath := getConfigPath()

	gitAdapter, err := gogit.NewAdapter(repoPath)
	if err != nil {
		log.Fatalf("Failed to open git repository: %v", err)
	}

	configAdapter, err := fileconfig.NewAdapter(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	model := ui.NewMainModel(gitAdapter, configAdapter)

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}

func getRepoPath() string {
	if path := os.Getenv("GITTO_REPO_PATH"); path != "" {
		return path
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get current working directory")
	}

	return cwd
}

func getConfigPath() string {
	if path := os.Getenv("GITTO_CONFIG_PATH"); path != "" {
		return path
	}

	if home, err := os.UserHomeDir(); err == nil {
		userConfig := fmt.Sprintf("%s/.config/gitto/config.toml", home)
		if _, err := os.Stat(userConfig); err == nil {
			return userConfig
		}
	}

	for _, p := range []string{"config/default.toml", ".gitto.toml"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return "config/default.toml"
}