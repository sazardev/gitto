package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/ports"
	"github.com/sazardev/gitto/internal/styles"
	"github.com/sazardev/gitto/internal/ui/views"
)

type ViewMode int

const (
	DashboardViewMode ViewMode = iota
	BranchesViewMode
	ChangesViewMode
	LogViewMode
	DiffViewMode
	CommitViewMode
	HelpViewMode
)

var tabNames = []string{"dashboard", "branches", "changes", "history"}

type MainModel struct {
	Git       ports.GitProvider
	Config    ports.ConfigProvider
	Linguist  ports.LanguageProvider
	Spinner   spinner.Model

	StatusView   views.StatusView
	BranchesView views.BranchesView
	ChangesView  views.ChangesView
	LogView      views.LogView
	DiffView     views.DiffView
	CommitView   views.CommitView

	Width  int
	Height int

	CurrentBranch string
	Files         []entities.FileStatus
	Language      entities.Language
	HasLanguage   bool
	ViewMode      ViewMode
	Loading       bool
	Err           error
	LastMessage   string
	ShowShortcuts bool
}

func NewMainModel(git ports.GitProvider, config ports.ConfigProvider, linguist ports.LanguageProvider) MainModel {
	return MainModel{
		Git:          git,
		Config:       config,
		Linguist:     linguist,
		Spinner:      spinner.New(spinner.WithSpinner(spinner.Line)),
		StatusView:   views.NewStatusView(),
		BranchesView: views.NewBranchesView(),
		ChangesView:  views.NewChangesView(),
		LogView:      views.NewLogView(),
		DiffView:     views.NewDiffView(),
		CommitView:   views.NewCommitView(),
		ViewMode:     DashboardViewMode,
		Loading:      false,
		ShowShortcuts: true,
	}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadDashboard(),
		m.detectLanguage(),
	)
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case spinner.TickMsg:
		s, cmd := m.Spinner.Update(msg)
		m.Spinner = s
		return m, cmd
	case CommitSuccess:
		m.ViewMode = DashboardViewMode
		m.CommitView.Hide()
		m.Loading = false
		m.LastMessage = "commit successful"
		return m, m.loadDashboard()
	case CommitError:
		m.Err = msg.Err
		m.Loading = false
		m.LastMessage = "commit failed: " + msg.Err.Error()
		return m, nil
	case PushSuccess:
		m.Loading = false
		m.LastMessage = "push successful"
		return m, m.loadDashboard()
	case PushError:
		m.Err = msg.Err
		m.Loading = false
		m.LastMessage = "push failed: " + msg.Err.Error()
		return m, nil
	case PullSuccess:
		m.Loading = false
		m.LastMessage = "pull successful"
		return m, m.loadDashboard()
	case PullError:
		m.Err = msg.Err
		m.Loading = false
		m.LastMessage = "pull failed: " + msg.Err.Error()
		return m, nil
	case CloseCommitView:
		m.ViewMode = DashboardViewMode
		m.CommitView.Hide()
		return m, nil
	case LogLoaded:
		m.LogView.Update(msg.Commits)
		return m, nil
	case LogError:
		m.Err = msg.Err
		m.LastMessage = "failed to load log: " + msg.Err.Error()
		return m, nil
	case DashboardLoaded:
		m.StatusView = m.StatusView.Update(msg.Dashboard.Files)
		m.StatusView = m.StatusView.UpdateBranches(msg.Dashboard.Branches)
		m.StatusView = m.StatusView.UpdateCommits(msg.Dashboard.Commits)
		m.BranchesView = m.BranchesView.Update(msg.Dashboard.Branches)
		m.ChangesView = m.ChangesView.Update(msg.Dashboard.Files)
		m.CurrentBranch = msg.Dashboard.CurrentBranch
		return m, nil
	case DashboardError:
		m.Err = msg.Err
		m.LastMessage = "failed to load dashboard: " + msg.Err.Error()
		return m, nil
	case LanguageDetected:
		m.Language = msg.Language
		m.HasLanguage = true
		styles.ApplyAccent(msg.Language.Color)
		return m, nil
	}
	return m, nil
}

func (m MainModel) View() string {
	w := m.Width
	h := m.Height
	if w == 0 { w = 80 }
	if h == 0 { h = 24 }

	if m.ViewMode == HelpViewMode {
		return m.renderHelpScreen(w, h)
	}

	contentHeight := h - styles.ChromeHeight
	if contentHeight < 3 {
		contentHeight = 3
	}

	var sb strings.Builder

	sb.WriteString(m.renderHeader(w))
	sb.WriteString("\n")
	sb.WriteString(m.renderTabBar(w))
	sb.WriteString("\n")

	switch m.ViewMode {
	case DashboardViewMode:
		sb.WriteString(m.StatusView.Render(w, contentHeight))
	case BranchesViewMode:
		sb.WriteString(m.BranchesView.Render(w, contentHeight))
	case ChangesViewMode:
		sb.WriteString(m.ChangesView.Render(w, contentHeight))
	case LogViewMode:
		sb.WriteString(m.LogView.Render(w, contentHeight))
	case DiffViewMode:
		sb.WriteString(m.DiffView.Render(w, contentHeight))
	case CommitViewMode:
		commitH := contentHeight / 3
		if commitH < 3 { commitH = 3 }
		sb.WriteString(m.StatusView.Render(w, contentHeight-commitH))
		sb.WriteString("\n")
		sb.WriteString(m.CommitView.Render(w, commitH))
	}

	sb.WriteString("\n")
	sb.WriteString(m.renderStatusBar(w))
	sb.WriteString("\n")
	sb.WriteString(m.renderFooter(w))

	return sb.String()
}

func (m MainModel) renderHeader(w int) string {
	branch := styles.BranchActiveStyle.Render(m.CurrentBranch)
	app := styles.HeaderBorderStyle.Render("gitto")

	var langInfo string
	if m.HasLanguage {
		langInfo = styles.DimStyle.Render(" · ") + styles.AccentStyle.Render(m.Language.Name)
	}

	gap := w - lipgloss.Width(app) - lipgloss.Width(branch) - lipgloss.Width(langInfo) - 4
	if gap < 1 { gap = 1 }
	pad := strings.Repeat(" ", gap)

	return app + pad + branch + langInfo
}

func (m MainModel) renderTabBar(w int) string {
	var parts []string

	for i, name := range tabNames {
		var tab string
		if ViewMode(i) == m.ViewMode {
			tab = styles.TabActiveStyle.Render(name)
		} else {
			tab = styles.TabInactiveStyle.Render(name)
		}
		num := styles.DimStyle.Render("[" + string(rune('1'+i)) + "] ")
		parts = append(parts, num+tab)
	}

	sep := styles.TabSeparatorStyle.Render("  ·  ")
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}

	return result
}

func (m MainModel) renderStatusBar(w int) string {
	if m.Loading {
		return styles.StatusLineStyle.Render(" " + m.Spinner.View() + " " + m.LastMessage)
	}
	if m.LastMessage != "" {
		msg := m.LastMessage
		if len(msg) > w-2 {
			msg = msg[:w-5] + "..."
		}
		return styles.StatusLineStyle.Render(" " + msg)
	}
	return styles.StatusLineStyle.Render(" " + m.ViewModeName())
}

func (m MainModel) renderFooter(w int) string {
	text := m.getHelpText(w)
	return styles.FooterBorderStyle.Render(text)
}

func (m MainModel) ViewModeName() string {
	switch m.ViewMode {
	case DashboardViewMode:
		return "dashboard"
	case BranchesViewMode:
		return "branches"
	case ChangesViewMode:
		return "changes"
	case LogViewMode:
		return "history"
	case DiffViewMode:
		return "diff"
	case CommitViewMode:
		return "commit"
	case HelpViewMode:
		return "help"
	default:
		return ""
	}
}

func (m MainModel) getHelpText(w int) string {
	if w < 50 {
		return "1-4 tabs  ↑↓ nav  [?] help  [q] quit"
	}
	if w < 70 {
		return "1-4 tabs  ↑↓ nav  [s] stage  [u] unstage  [?] help  [/] bar  [q] quit"
	}
	if w < 100 {
		return "1-4 tabs  ↑↓ nav  [s] stage  [u] unstage  [d] diff  [c] commit  [?] help  [q] quit"
	}
	return "1-4 tabs  ↑↓ nav  [s] stage  [u] unstage  [d] diff  [c] commit  [P] push  [p] pull  [?] help  [q] quit"
}

func (m MainModel) renderHelpScreen(width, height int) string {
	title := styles.TitleStyle.Render("?  keyboard shortcuts")

	sections := []struct {
		name  string
		items []struct{ key, desc string }
	}{
		{
			name: "navigation",
			items: []struct{ key, desc string }{
				{"1", "dashboard view"},
				{"2", "branches view"},
				{"3", "changes view"},
				{"4", "history view"},
				{"tab", "next tab"},
				{"shift+tab", "previous tab"},
				{"↑/↓", "move up/down"},
			},
		},
		{
			name: "git operations",
			items: []struct{ key, desc string }{
				{"s", "stage selected file"},
				{"u", "unstage selected file"},
				{"d", "view diff of selected file"},
				{"c", "open commit input"},
				{"P", "push to remote"},
				{"p", "pull from remote"},
				{"r", "reload dashboard"},
			},
		},
		{
			name: "general",
			items: []struct{ key, desc string }{
				{"?", "show this help screen"},
				{"/", "toggle shortcuts bar"},
				{"esc", "go back / close"},
				{"q", "quit"},
			},
		},
	}

	keyStyle := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true).Width(14)
	descStyle := lipgloss.NewStyle().Foreground(styles.Foreground)
	sectionStyle := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true).Underline(true)

	var lines []string
	lines = append(lines, "")
	lines = append(lines, "  "+title)
	lines = append(lines, "")

	for _, section := range sections {
		lines = append(lines, "  "+sectionStyle.Render(section.name))
		for _, item := range section.items {
			lines = append(lines, "    "+keyStyle.Render("["+item.key+"]")+"  "+descStyle.Render(item.desc))
		}
		lines = append(lines, "")
	}

	lines = append(lines, "  "+styles.HelpStyle.Render("press [?] or [esc] to close"))

	content := ""
	for _, line := range lines {
		content += line + "\n"
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		Padding(1, 2).
		Width(min(width-4, 60)).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func (m MainModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ViewMode == CommitViewMode {
		return m.handleCommitKey(msg)
	}

	if m.ViewMode == HelpViewMode {
		switch msg.String() {
		case "?", "esc", "q":
			m.ViewMode = DashboardViewMode
		}
		return m, nil
	}

	switch msg.Type {
	case tea.KeyRunes:
		switch msg.String() {
		case "1":
			m.ViewMode = DashboardViewMode
			return m, nil
		case "2":
			m.ViewMode = BranchesViewMode
			return m, nil
		case "3":
			m.ViewMode = ChangesViewMode
			return m, nil
		case "4":
			m.ViewMode = LogViewMode
			return m, m.loadLog()
		case "r":
			return m, m.loadDashboard()
		case "d":
			if m.ViewMode == ChangesViewMode {
				if f := m.ChangesView.SelectedFile(); f != nil {
					m.ViewMode = DiffViewMode
					return m, m.loadDiff(f.Path, f.IsStaged)
				}
			} else if f := m.StatusView.SelectedFile(); f != nil {
				m.ViewMode = DiffViewMode
				return m, m.loadDiff(f.Path, f.IsStaged)
			}
		case "c":
			m.ViewMode = CommitViewMode
			m.CommitView.Show()
		case "s":
			if m.ViewMode == ChangesViewMode {
				if f := m.ChangesView.SelectedFile(); f != nil {
					return m, m.stageFile(f.Path)
				}
			} else if f := m.StatusView.SelectedFile(); f != nil {
				return m, m.stageFile(f.Path)
			}
		case "u":
			if m.ViewMode == ChangesViewMode {
				if f := m.ChangesView.SelectedFile(); f != nil {
					return m, m.unstageFile(f.Path)
				}
			} else if f := m.StatusView.SelectedFile(); f != nil {
				return m, m.unstageFile(f.Path)
			}
		case "P":
			return m, m.push()
		case "p":
			return m, m.pull()
		case "q":
			return m, tea.Quit
		case "?":
			m.ViewMode = HelpViewMode
			return m, nil
		case "/":
			m.ShowShortcuts = !m.ShowShortcuts
			return m, nil
		}
	case tea.KeyTab:
		m.ViewMode = (m.ViewMode + 1) % ViewMode(len(tabNames))
		if m.ViewMode == LogViewMode {
			return m, m.loadLog()
		}
		return m, nil
	case tea.KeyShiftTab:
		m.ViewMode = (m.ViewMode - 1 + ViewMode(len(tabNames))) % ViewMode(len(tabNames))
		if m.ViewMode == LogViewMode {
			return m, m.loadLog()
		}
		return m, nil
	case tea.KeyUp:
		switch m.ViewMode {
		case DashboardViewMode:
			m.StatusView.MoveUp()
		case BranchesViewMode:
			m.BranchesView.MoveUp()
		case ChangesViewMode:
			m.ChangesView.MoveUp()
		case LogViewMode:
			m.LogView.MoveUp()
		}
	case tea.KeyDown:
		switch m.ViewMode {
		case DashboardViewMode:
			m.StatusView.MoveDown()
		case BranchesViewMode:
			m.BranchesView.MoveDown()
		case ChangesViewMode:
			m.ChangesView.MoveDown()
		case LogViewMode:
			m.LogView.MoveDown()
		}
	case tea.KeyEsc:
		if m.ViewMode == DiffViewMode {
			m.ViewMode = ChangesViewMode
		} else if m.ViewMode != DashboardViewMode {
			m.ViewMode = DashboardViewMode
		}
	}

	return m, nil
}

func (m MainModel) handleCommitKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		return m, m.handleCommitSubmit()
	case tea.KeyEsc:
		m.ViewMode = DashboardViewMode
		m.CommitView.Hide()
		return m, nil
	}

	var cmd tea.Cmd
	m.CommitView.Input, cmd = m.CommitView.Input.Update(msg)
	return m, cmd
}

func (m MainModel) handleCommitSubmit() tea.Cmd {
	if m.CommitView.Input.Value() == "" {
		return nil
	}
	m.Loading = true
	return func() tea.Msg {
		err := m.Git.Commit(m.CommitView.Input.Value())
		if err != nil {
			return CommitError{Err: err}
		}
		return CommitSuccess{}
	}
}

func (m MainModel) loadDashboard() tea.Cmd {
	return func() tea.Msg {
		dashboard := entities.NewDashboard()

		branch, err := m.Git.GetCurrentBranch()
		if err != nil {
			return DashboardError{Err: err}
		}
		dashboard.CurrentBranch = branch

		branches, err := m.Git.GetBranches()
		if err != nil {
			return DashboardError{Err: err}
		}
		dashboard.Branches = branches

		commits, err := m.Git.GetLog(m.Config.GetMaxLogItems())
		if err != nil {
			return DashboardError{Err: err}
		}
		dashboard.Commits = commits

		files, err := m.Git.GetStatus()
		if err != nil {
			return DashboardError{Err: err}
		}
		dashboard.Files = files

		for _, f := range files {
			if f.IsUntracked {
				dashboard.UntrackedCount++
			} else if f.IsStaged {
				dashboard.StagedCount++
			} else {
				dashboard.UnstagedCount++
			}
		}

		return DashboardLoaded{Dashboard: dashboard}
	}
}

func (m MainModel) detectLanguage() tea.Cmd {
	return func() tea.Msg {
		if m.Linguist == nil {
			return nil
		}
		lang, ok := m.Linguist.DetectLanguage()
		if !ok {
			return nil
		}
		return LanguageDetected{Language: lang}
	}
}

func (m MainModel) loadLog() tea.Cmd {
	return func() tea.Msg {
		commits, err := m.Git.GetLog(m.Config.GetMaxLogItems())
		if err != nil {
			return LogError{Err: err}
		}
		return LogLoaded{Commits: commits}
	}
}

func (m MainModel) loadDiff(path string, staged bool) tea.Cmd {
	return func() tea.Msg {
		diff, err := m.Git.GetDiff(path, staged)
		if err != nil {
			return err
		}
		m.DiffView.Update(diff)
		return nil
	}
}

func (m MainModel) stageFile(path string) tea.Cmd {
	return func() tea.Msg {
		err := m.Git.Stage([]string{path})
		if err != nil {
			return err
		}
		return m.loadDashboard()()
	}
}

func (m MainModel) unstageFile(path string) tea.Cmd {
	return func() tea.Msg {
		err := m.Git.Unstage([]string{path})
		if err != nil {
			return err
		}
		return m.loadDashboard()()
	}
}

func (m MainModel) push() tea.Cmd {
	m.Loading = true
	m.LastMessage = "pushing..."
	return func() tea.Msg {
		err := m.Git.Push()
		if err != nil {
			return PushError{Err: err}
		}
		return PushSuccess{}
	}
}

func (m MainModel) pull() tea.Cmd {
	m.Loading = true
	m.LastMessage = "pulling..."
	return func() tea.Msg {
		err := m.Git.Pull()
		if err != nil {
			return PullError{Err: err}
		}
		return PullSuccess{}
	}
}

var _ tea.Model = (*MainModel)(nil)
