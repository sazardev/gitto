package views

import (
	"strings"

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
	ti.Placeholder = "commit message..."
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

func (v CommitView) Render(width, height int) string {
	if !v.Visible {
		return ""
	}

	inputWidth := width - 8
	if inputWidth < 20 {
		inputWidth = 20
	}
	v.Input.Width = inputWidth

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(styles.PanelTitleStyle.Render("commit"))
	sb.WriteString("\n")
	sb.WriteString(v.Input.View())
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(styles.Dim).Render("press enter to commit, esc to cancel"))

	return sb.String()
}
