package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/ports"
	"github.com/sazardev/gitto/internal/ui/views"
)

var (
	appStyle = lipgloss.NewStyle()

	headerBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Accent).
				Padding(0, 1)

	helpBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Dim).
				Foreground(Dim).
				Padding(0, 1)
)

type ViewMode int

const (
	StatusViewMode ViewMode = iota
	LogViewMode
	DiffViewMode
	CommitViewMode
)

type MainModel struct {
	Git      ports.GitProvider
	Config   ports.ConfigProvider
	Spinner  spinner.Model

	StatusView  views.StatusView
	LogView     views.LogView
	DiffView    views.DiffView
	CommitView  views.CommitView

	CurrentBranch string
	Files         []entities.FileStatus
	ViewMode      ViewMode
	Loading       bool
	Err           error
	LastMessage   string
}

func NewMainModel(git ports.GitProvider, config ports.ConfigProvider) MainModel {
	return MainModel{
		Git:          git,
		Config:       config,
		Spinner:      spinner.New(spinner.WithSpinner(spinner.Line)),
		StatusView:   views.NewStatusView(),
		LogView:      views.NewLogView(),
		DiffView:     views.NewDiffView(),
		CommitView:   views.NewCommitView(),
		ViewMode:     StatusViewMode,
		Loading:      false,
	}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadDashboard(),
	)
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case spinner.TickMsg:
		s, cmd := m.Spinner.Update(msg)
		m.Spinner = s
		return m, cmd
	case CommitSuccess:
		m.ViewMode = StatusViewMode
		m.CommitView.Hide()
		m.Loading = false
		m.LastMessage = "Commit successful"
		return m, m.loadDashboard()
	case CommitError:
		m.Err = msg.Err
		m.Loading = false
		m.LastMessage = "Commit failed: " + msg.Err.Error()
		return m, nil
	case PushSuccess:
		m.Loading = false
		m.LastMessage = "Push successful"
		return m, m.loadDashboard()
	case PushError:
		m.Err = msg.Err
		m.Loading = false
		m.LastMessage = "Push failed: " + msg.Err.Error()
		return m, nil
	case PullSuccess:
		m.Loading = false
		m.LastMessage = "Pull successful"
		return m, m.loadDashboard()
	case PullError:
		m.Err = msg.Err
		m.Loading = false
		m.LastMessage = "Pull failed: " + msg.Err.Error()
		return m, nil
	case CloseCommitView:
		m.ViewMode = StatusViewMode
		m.CommitView.Hide()
		return m, nil
	case LogLoaded:
		m.LogView.Update(msg.Commits)
		return m, nil
	case LogError:
		m.Err = msg.Err
		m.LastMessage = "Failed to load log: " + msg.Err.Error()
		return m, nil
	case DashboardLoaded:
		m.StatusView = m.StatusView.Update(msg.Dashboard.Files)
		m.StatusView = m.StatusView.UpdateBranches(msg.Dashboard.Branches)
		m.StatusView = m.StatusView.UpdateCommits(msg.Dashboard.Commits)
		m.CurrentBranch = msg.Dashboard.CurrentBranch
		return m, nil
	case DashboardError:
		m.Err = msg.Err
		m.LastMessage = "Failed to load dashboard: " + msg.Err.Error()
		return m, nil
	}
	return m, nil
}

func (m MainModel) View() string {
	header := headerBorderStyle.Render("[ gitto ]") + "  " + BranchStyle.Render(m.CurrentBranch) + "\n"

	var viewContent string
	if m.Loading {
		viewContent = m.Spinner.View() + " " + m.LastMessage + "...\n\n"
	}

	switch m.ViewMode {
	case StatusViewMode:
		viewContent += m.StatusView.Render()
	case LogViewMode:
		viewContent += m.LogView.Render()
	case DiffViewMode:
		viewContent += m.DiffView.Render()
	case CommitViewMode:
		viewContent += m.StatusView.Render()
		viewContent += "\n"
		viewContent += m.CommitView.Render()
	}

	if m.LastMessage != "" && !m.Loading {
		viewContent += "\n"
		viewContent += m.LastMessage
	}

	help := helpBorderStyle.Render("[r] Refresh  [l] Log  [c] Commit  [d] Diff  [s] Stage  [u] Unstage  [P] Push  [p] Pull  [q] Quit")

	fullView := header + "\n" + viewContent + "\n" + help
	return appStyle.Render(fullView)
}

func (m MainModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ViewMode == CommitViewMode {
		return m.handleCommitKey(msg)
	}

	switch msg.Type {
	case tea.KeyRunes:
		switch msg.String() {
		case "r":
			m.ViewMode = StatusViewMode
			return m, m.loadDashboard()
		case "l":
			m.ViewMode = LogViewMode
			return m, m.loadLog()
		case "d":
			if f := m.StatusView.SelectedFile(); f != nil {
				m.ViewMode = DiffViewMode
				return m, m.loadDiff(f.Path, f.IsStaged)
			}
		case "c":
			m.ViewMode = CommitViewMode
			m.CommitView.Show()
		case "s":
			if f := m.StatusView.SelectedFile(); f != nil {
				return m, m.stageFile(f.Path)
			}
		case "u":
			if f := m.StatusView.SelectedFile(); f != nil {
				return m, m.unstageFile(f.Path)
			}
		case "P":
			return m, m.push()
		case "p":
			return m, m.pull()
		case "q":
			return m, tea.Quit
		}
	case tea.KeyUp:
		switch m.ViewMode {
		case StatusViewMode:
			m.StatusView.MoveUp()
		case LogViewMode:
			m.LogView.MoveUp()
		}
	case tea.KeyDown:
		switch m.ViewMode {
		case StatusViewMode:
			m.StatusView.MoveDown()
		case LogViewMode:
			m.LogView.MoveDown()
		}
	case tea.KeyEsc:
		if m.ViewMode != CommitViewMode {
			m.ViewMode = StatusViewMode
		}
	}

	return m, nil
}

func (m MainModel) handleCommitKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		return m, m.handleCommitSubmit()
	case tea.KeyEsc:
		m.ViewMode = StatusViewMode
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

func (m MainModel) loadStatus() tea.Cmd {
	return func() tea.Msg {
		files, err := m.Git.GetStatus()
		if err != nil {
			return err
		}
		m.Files = files
		m.StatusView = m.StatusView.Update(files)
		return nil
	}
}

func (m MainModel) loadBranch() tea.Cmd {
	return func() tea.Msg {
		branch, err := m.Git.GetCurrentBranch()
		if err != nil {
			return err
		}
		m.CurrentBranch = branch
		return nil
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
		files, err := m.Git.GetStatus()
		if err != nil {
			return err
		}
		m.Files = files
		m.StatusView = m.StatusView.Update(files)
		return nil
	}
}

func (m MainModel) unstageFile(path string) tea.Cmd {
	return func() tea.Msg {
		err := m.Git.Unstage([]string{path})
		if err != nil {
			return err
		}
		files, err := m.Git.GetStatus()
		if err != nil {
			return err
		}
		m.Files = files
		m.StatusView = m.StatusView.Update(files)
		return nil
	}
}

func (m MainModel) push() tea.Cmd {
	m.Loading = true
	m.LastMessage = "Pushing..."
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
	m.LastMessage = "Pulling..."
	return func() tea.Msg {
		err := m.Git.Pull()
		if err != nil {
			return PullError{Err: err}
		}
		return PullSuccess{}
	}
}

var _ tea.Model = (*MainModel)(nil)
