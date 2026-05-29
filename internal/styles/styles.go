package styles

import "github.com/charmbracelet/lipgloss"

var (
	Background     = lipgloss.Color("#1a1a2e")
	Foreground     = lipgloss.Color("#eaeaea")
	Accent         = lipgloss.Color("#7c3aed")
	Success        = lipgloss.Color("#22c55e")
	Warning        = lipgloss.Color("#f59e0b")
	Error          = lipgloss.Color("#ef4444")
	DiffAdded      = lipgloss.Color("#22c55e")
	DiffRemoved    = lipgloss.Color("#ef4444")
	Dim            = lipgloss.Color("#6b7280")

	StatusStyle    = lipgloss.NewStyle().Background(Background).Foreground(Foreground)

	BranchStyle    = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	StagedStyle    = lipgloss.NewStyle().Foreground(Success)

	UnstagedStyle  = lipgloss.NewStyle().Foreground(Warning)

	UntrackedStyle  = lipgloss.NewStyle().Foreground(Dim)

	SelectedStyle  = lipgloss.NewStyle().Background(Accent).Foreground(Foreground)

	TitleStyle     = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	BorderStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Accent)

	SpinnerStyle   = lipgloss.NewStyle().Foreground(Accent)

	HelpStyle      = lipgloss.NewStyle().Foreground(Dim)

	DiffAddedStyle  = lipgloss.NewStyle().Foreground(DiffAdded)
	DiffRemovedStyle = lipgloss.NewStyle().Foreground(DiffRemoved)
	AccentStyle     = lipgloss.NewStyle().Foreground(Accent)
)