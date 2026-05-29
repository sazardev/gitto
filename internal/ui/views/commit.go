package views

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/sazardev/gitto/internal/styles"
)

type CommitView struct {
	Input   textinput.Model
	Visible bool
}

func NewCommitView() CommitView {
	ti := textinput.New()
	ti.Placeholder = "Commit message..."
	ti.Focus()

	return CommitView{
		Input:   ti,
		Visible: false,
	}
}

func (v *CommitView) Show() {
	v.Visible = true
	v.Input.Focus()
}

func (v *CommitView) Hide() {
	v.Visible = false
	v.Input.Reset()
}

func (v CommitView) Render() string {
	if !v.Visible {
		return ""
	}

	s := "\n\n"
	s += styles.TitleStyle.Render("Commit") + "\n\n"
	s += v.Input.View() + "\n"
	s += lipgloss.NewStyle().Foreground(styles.Dim).Render("Press Enter to commit, Esc to cancel")

	return s
}