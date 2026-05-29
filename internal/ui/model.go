package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/ports"
	"github.com/sazardev/gitto/internal/ui/views"
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
		m.loadStatus(),
		m.loadBranch(),
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
		return m, m.loadStatus()
	case CommitError:
		m.Err = msg.Err
		m.Loading = false
		m.LastMessage = "Commit failed: " + msg.Err.Error()
		return m, nil
	case PushSuccess:
		m.Loading = false
		m.LastMessage = "Push successful"
		return m, m.loadStatus()
	case PushError:
		m.Err = msg.Err
		m.Loading = false
		m.LastMessage = "Push failed: " + msg.Err.Error()
		return m, nil
	case PullSuccess:
		m.Loading = false
		m.LastMessage = "Pull successful"
		return m, m.loadStatus()
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
	}
	return m, nil
}

func (m MainModel) View() string {
	var s strings.Builder

	if m.Loading {
		s.WriteString(m.Spinner.View())
		s.WriteString(" " + m.LastMessage + "...\n\n")
	}

	switch m.ViewMode {
	case StatusViewMode:
		s.WriteString(m.StatusView.Render())
	case LogViewMode:
		s.WriteString(m.LogView.Render())
	case DiffViewMode:
		s.WriteString(m.DiffView.Render())
	case CommitViewMode:
		s.WriteString(m.StatusView.Render())
		s.WriteString("\n")
		s.WriteString(m.CommitView.Render())
	}

	if m.LastMessage != "" {
		s.WriteString("\n")
		s.WriteString(m.LastMessage)
	}

	return s.String()
}

func (m MainModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ViewMode == CommitViewMode {
		return m.handleCommitKey(msg)
	}

	switch {
	case msg.Type == tea.KeyRunes:
		switch msg.String() {
		case "r":
			m.ViewMode = StatusViewMode
			return m, m.loadStatus()
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
		}
	case msg.Type == tea.KeyUp:
		if m.ViewMode == LogViewMode {
			m.LogView.MoveUp()
			return m, nil
		}
	case msg.Type == tea.KeyDown:
		if m.ViewMode == LogViewMode {
			m.LogView.MoveDown()
			return m, nil
		}
	case msg.Type == tea.KeyEsc:
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
		m.StatusView.Update(files)
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
		m.StatusView.Update(files)
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
		m.StatusView.Update(files)
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
