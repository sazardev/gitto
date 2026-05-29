package components

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func NewFuzzyList(items []list.Item) list.Model {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	return l
}

var listStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#1a1a2e")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#7c3aed"))