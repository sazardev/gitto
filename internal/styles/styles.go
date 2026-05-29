package styles

import "github.com/charmbracelet/lipgloss"

var (
	Background     = lipgloss.Color("")
	Foreground     = lipgloss.Color("")
	Accent         = lipgloss.Color("6")
	Success        = lipgloss.Color("2")
	Warning        = lipgloss.Color("3")
	Error          = lipgloss.Color("1")
	DiffAdded      = lipgloss.Color("2")
	DiffRemoved    = lipgloss.Color("1")
	Dim            = lipgloss.Color("8")

	StatusStyle    = lipgloss.NewStyle()

	BranchStyle    = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	StagedStyle    = lipgloss.NewStyle().Foreground(Success)

	UnstagedStyle  = lipgloss.NewStyle().Foreground(Warning)

	UntrackedStyle  = lipgloss.NewStyle().Foreground(Dim)

	SelectedStyle  = lipgloss.NewStyle().Foreground(Accent)

	TitleStyle     = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	BorderStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Accent)

	SpinnerStyle   = lipgloss.NewStyle().Foreground(Accent)

	HelpStyle      = lipgloss.NewStyle().Foreground(Dim)
	DimStyle       = lipgloss.NewStyle().Foreground(Dim)
	BoldStyle      = lipgloss.NewStyle().Bold(true)

	DiffAddedStyle  = lipgloss.NewStyle().Foreground(DiffAdded)
	DiffRemovedStyle = lipgloss.NewStyle().Foreground(DiffRemoved)
	AccentStyle     = lipgloss.NewStyle().Foreground(Accent)
	HashStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	TimeStyle       = lipgloss.NewStyle().Foreground(Dim)

	PanelStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Accent).Padding(0, 1)
	PanelTitleStyle = lipgloss.NewStyle().Foreground(Accent).Bold(true).Underline(true)

	BranchActiveStyle = lipgloss.NewStyle().Foreground(Success).Bold(true)
	BranchItemStyle   = lipgloss.NewStyle().Foreground(Foreground)
	BranchRemoteStyle = lipgloss.NewStyle().Foreground(Dim)

	CommitHashStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	CommitAuthorStyle = lipgloss.NewStyle().Foreground(Dim)
	CommitTimeStyle   = lipgloss.NewStyle().Foreground(Accent)
	CommitMsgStyle    = lipgloss.NewStyle().Foreground(Foreground)

	StatStagedStyle   = lipgloss.NewStyle().Foreground(Success).Bold(true)
	StatUnstagedStyle = lipgloss.NewStyle().Foreground(Warning).Bold(true)
	StatUntrackedStyle = lipgloss.NewStyle().Foreground(Dim).Bold(true)
)