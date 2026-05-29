package ui

import "github.com/charmbracelet/lipgloss"

var (
	Background = lipgloss.Color("") 
	Foreground = lipgloss.Color("") 
	Accent     = lipgloss.Color("6") 
	Success    = lipgloss.Color("2") 
	Warning    = lipgloss.Color("3") 
	Error      = lipgloss.Color("1") 
	DiffAdded  = lipgloss.Color("2") 
	DiffRemoved = lipgloss.Color("1") 
	Dim        = lipgloss.Color("8") 

	StatusStyle = lipgloss.NewStyle()

	BranchStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	StagedStyle = lipgloss.NewStyle().
			Foreground(Success)

	UnstagedStyle = lipgloss.NewStyle().
			Foreground(Warning)

	UntrackedStyle = lipgloss.NewStyle().
			Foreground(Dim)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(Accent)

	TitleStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Accent)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Accent)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Dim)
)