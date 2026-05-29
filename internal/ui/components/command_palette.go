package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type CommandPalette struct {
	Input      textinput.Model
	Visible    bool
	Results    []string
	Selected   int
	onSubmit   func(string)
}

func NewCommandPalette(onSubmit func(string)) CommandPalette {
	ti := textinput.New()
	ti.Placeholder = ":"
	ti.Prompt = ":"
	ti.Focus()

	return CommandPalette{
		Input:    ti,
		Visible:  false,
		Results:  []string{},
		Selected: 0,
		onSubmit: onSubmit,
	}
}

func (p *CommandPalette) Show() {
	p.Visible = true
	p.Input.Focus()
}

func (p *CommandPalette) Hide() {
	p.Visible = false
	p.Input.Reset()
}

func (p *CommandPalette) Toggle() {
	if p.Visible {
		p.Hide()
	} else {
		p.Show()
	}
}

var paletteStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#1a1a2e")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#7c3aed")).
	Padding(1)