package styles

import "github.com/charmbracelet/lipgloss"

const (
	HeaderHeight  = 1
	TabBarHeight  = 1
	FooterHeight  = 1
	StatusHeight  = 1
	ChromeHeight  = HeaderHeight + TabBarHeight + FooterHeight + StatusHeight
)

var (
	Background = lipgloss.Color("")
	Foreground = lipgloss.Color("")
	Accent     = lipgloss.Color("6")
	Success    = lipgloss.Color("2")
	Warning    = lipgloss.Color("3")
	Error      = lipgloss.Color("1")
	Dim        = lipgloss.Color("8")
	Border     = lipgloss.Color("8")
	Muted      = lipgloss.Color("240")
)

var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	DimStyle = lipgloss.NewStyle().
			Foreground(Dim)

	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)
)

var (
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1)

	PanelTitleStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true).
			PaddingBottom(0)
)

var (
	BranchActiveStyle = lipgloss.NewStyle().
				Foreground(Success).
				Bold(true)

	BranchItemStyle = lipgloss.NewStyle().
			Foreground(Foreground)

	BranchRemoteStyle = lipgloss.NewStyle().
				Foreground(Dim)
)

var (
	CommitHashStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	CommitAuthorStyle = lipgloss.NewStyle().Foreground(Dim)
	CommitTimeStyle   = lipgloss.NewStyle().Foreground(Accent)
	CommitMsgStyle    = lipgloss.NewStyle().Foreground(Foreground)
)

var (
	StagedStyle    = lipgloss.NewStyle().Foreground(Success)
	UnstagedStyle  = lipgloss.NewStyle().Foreground(Warning)
	UntrackedStyle = lipgloss.NewStyle().Foreground(Dim)
	SelectedStyle  = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	StatStagedStyle    = lipgloss.NewStyle().Foreground(Success).Bold(true)
	StatUnstagedStyle  = lipgloss.NewStyle().Foreground(Warning).Bold(true)
	StatUntrackedStyle = lipgloss.NewStyle().Foreground(Dim).Bold(true)
)

var (
	HashStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	TimeStyle    = lipgloss.NewStyle().Foreground(Dim)
	AccentStyle  = lipgloss.NewStyle().Foreground(Accent)
)

var (
	DiffAddedStyle   = lipgloss.NewStyle().Foreground(Success)
	DiffRemovedStyle = lipgloss.NewStyle().Foreground(Error)
	DiffHeaderStyle  = lipgloss.NewStyle().Foreground(Accent)
)

var (
	SpinnerStyle = lipgloss.NewStyle().Foreground(Accent)
	HelpStyle    = lipgloss.NewStyle().Foreground(Dim)
)

var (
	TabActiveStyle    = lipgloss.NewStyle().Foreground(Accent).Bold(true).Underline(true)
	TabInactiveStyle  = lipgloss.NewStyle().Foreground(Dim)
	TabSeparatorStyle = lipgloss.NewStyle().Foreground(Border)
)

var (
	HeaderBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Accent).
				Padding(0, 1)

	FooterBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Border).
				Padding(0, 1)

	StatusLineStyle = lipgloss.NewStyle().
			Foreground(Dim).
			Faint(true)
)

var (
	HorizontalLine = lipgloss.NewStyle().
			Foreground(Border)
)

func HLine(width int) string {
	r := ""
	for i := 0; i < width; i++ {
		r += "─"
	}
	return HorizontalLine.Render(r)
}

func Separator(width int) string {
	return HLine(width)
}

func ApplyAccent(color string) {
	Accent = lipgloss.Color(color)
	Border = lipgloss.Color(color)

	TitleStyle = TitleStyle.Foreground(Accent)
	PanelTitleStyle = PanelTitleStyle.Foreground(Accent)
	SelectedStyle = SelectedStyle.Foreground(Accent)
	AccentStyle = AccentStyle.Foreground(Accent)
	DiffHeaderStyle = DiffHeaderStyle.Foreground(Accent)
	SpinnerStyle = SpinnerStyle.Foreground(Accent)
	CommitTimeStyle = CommitTimeStyle.Foreground(Accent)
	TabActiveStyle = TabActiveStyle.Foreground(Accent)
	PanelStyle = PanelStyle.BorderForeground(Accent)
	HeaderBorderStyle = HeaderBorderStyle.BorderForeground(Accent)
}
