package ui

import (
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
	}
	return m, nil
}

func (m MainModel) View() string {
	if m.Loading {
		return m.Spinner.View() + " Loading..."
	}

	switch m.ViewMode {
	case StatusViewMode:
		return m.StatusView.Render()
	case LogViewMode:
		return m.LogView.Render()
	case DiffViewMode:
		return m.DiffView.Render()
	}

	return ""
}

func (m MainModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
				return m, m.loadDiff(f.Path)
			}
		case "c":
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
	case msg.Type == tea.KeyEsc:
		m.ViewMode = StatusViewMode
	}

	return m, nil
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
			return err
		}
		m.LogView.Update(commits)
		return nil
	}
}

func (m MainModel) loadDiff(path string) tea.Cmd {
	return func() tea.Msg {
		diff, err := m.Git.GetDiff(path)
		if err != nil {
			return err
		}
		m.DiffView.Update(diff, path)
		return nil
	}
}

func (m MainModel) stageFile(path string) tea.Cmd {
	return func() tea.Msg {
		err := m.Git.Stage([]string{path})
		if err != nil {
			return err
		}
		return m.loadStatus()
	}
}

func (m MainModel) unstageFile(path string) tea.Cmd {
	return func() tea.Msg {
		err := m.Git.Unstage([]string{path})
		if err != nil {
			return err
		}
		return m.loadStatus()
	}
}

func (m MainModel) push() tea.Cmd {
	return func() tea.Msg {
		return m.Git.Push()
	}
}

func (m MainModel) pull() tea.Cmd {
	return func() tea.Msg {
		return m.Git.Pull()
	}
}

var _ tea.Model = (*MainModel)(nil)