package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/styles"
)

type StatusView struct {
	Staged     []entities.FileStatus
	Unstaged   []entities.FileStatus
	Untracked  []entities.FileStatus
	Branches   []entities.Branch
	Commits    []entities.Commit
	Selected   int
	Height     int
	StagedCount   int
	UnstagedCount int
	UntrackedCount int
}

func NewStatusView() StatusView {
	return StatusView{
		Staged:    []entities.FileStatus{},
		Unstaged:  []entities.FileStatus{},
		Untracked: []entities.FileStatus{},
		Branches:  []entities.Branch{},
		Commits:   []entities.Commit{},
		Selected:  0,
	}
}

func (v StatusView) Update(files []entities.FileStatus) StatusView {
	v.Staged = []entities.FileStatus{}
	v.Unstaged = []entities.FileStatus{}
	v.Untracked = []entities.FileStatus{}

	for _, f := range files {
		if f.IsUntracked {
			v.Untracked = append(v.Untracked, f)
		} else if f.IsStaged {
			v.Staged = append(v.Staged, f)
		} else {
			v.Unstaged = append(v.Unstaged, f)
		}
	}

	v.StagedCount = len(v.Staged)
	v.UnstagedCount = len(v.Unstaged)
	v.UntrackedCount = len(v.Untracked)

	return v
}

func (v StatusView) UpdateBranches(branches []entities.Branch) StatusView {
	v.Branches = branches
	return v
}

func (v StatusView) UpdateCommits(commits []entities.Commit) StatusView {
	v.Commits = commits
	return v
}

func (v StatusView) Render() string {
	var s strings.Builder

	s.WriteString(v.renderBranchesPanel())
	s.WriteString("\n")
	s.WriteString(v.renderCommitsPanel())
	s.WriteString("\n")
	s.WriteString(v.renderChangesPanel())

	return s.String()
}

func (v StatusView) renderBranchesPanel() string {
	var s strings.Builder

	s.WriteString(styles.PanelTitleStyle.Render("Branches"))
	s.WriteString("\n")

	if len(v.Branches) == 0 {
		s.WriteString(styles.DimStyle.Render("  No branches"))
		return s.String()
	}

	for _, b := range v.Branches {
		var prefix string
		var style lipgloss.Style

		if b.IsHead {
			prefix = "> "
			style = styles.BranchActiveStyle
		} else if b.IsRemote {
			prefix = "  "
			style = styles.BranchRemoteStyle
		} else {
			prefix = "  "
			style = styles.BranchItemStyle
		}

		s.WriteString(prefix + style.Render(b.Name))
		s.WriteString("\n")
	}

	return styles.PanelStyle.Render(s.String())
}

func (v StatusView) renderCommitsPanel() string {
	var s strings.Builder

	s.WriteString(styles.PanelTitleStyle.Render("Recent Commits"))
	s.WriteString("\n")

	if len(v.Commits) == 0 {
		s.WriteString(styles.DimStyle.Render("  No commits found"))
		return styles.PanelStyle.Render(s.String())
	}

	maxCommits := 6
	if len(v.Commits) < maxCommits {
		maxCommits = len(v.Commits)
	}

	for i := 0; i < maxCommits; i++ {
		c := v.Commits[i]
		s.WriteString(styles.CommitHashStyle.Render(c.ShortHash))
		s.WriteString(" ")
		s.WriteString(styles.CommitAuthorStyle.Render(truncateStr(c.Author, 12)))
		s.WriteString(" ")
		s.WriteString(styles.CommitTimeStyle.Render(formatTime(c.AuthorDate)))
		s.WriteString("  ")
		s.WriteString(styles.CommitMsgStyle.Render(truncateStr(c.Message, 35)))
		s.WriteString("\n")
	}

	return styles.PanelStyle.Render(s.String())
}

func (v StatusView) renderChangesPanel() string {
	var s strings.Builder

	s.WriteString(styles.PanelTitleStyle.Render("Changes"))
	s.WriteString("\n")

	if v.Total() == 0 {
		s.WriteString(styles.DimStyle.Render("  No changes"))
		return styles.PanelStyle.Render(s.String())
	}

	statsLine := fmt.Sprintf("%s staged  %s unstaged  %s untracked",
		styles.StatStagedStyle.Render(fmt.Sprintf("%d", v.StagedCount)),
		styles.StatUnstagedStyle.Render(fmt.Sprintf("%d", v.UnstagedCount)),
		styles.StatUntrackedStyle.Render(fmt.Sprintf("%d", v.UntrackedCount)),
	)
	s.WriteString(statsLine)
	s.WriteString("\n\n")

	if len(v.Staged) > 0 {
		s.WriteString(styles.StagedStyle.Render("Staged:"))
		s.WriteString("\n")
		for i, f := range v.Staged {
			prefix := "  "
			if i == v.Selected {
				prefix = styles.SelectedStyle.Render("> ")
			}
			s.WriteString(prefix + f.Path + "\n")
		}
	}

	if len(v.Unstaged) > 0 {
		s.WriteString(styles.UnstagedStyle.Render("Unstaged:"))
		s.WriteString("\n")
		for i, f := range v.Unstaged {
			prefix := "  "
			if i+v.LenStaged() == v.Selected {
				prefix = styles.SelectedStyle.Render("> ")
			}
			s.WriteString(prefix + f.Path + "\n")
		}
	}

	if len(v.Untracked) > 0 {
		s.WriteString(styles.UntrackedStyle.Render("Untracked:"))
		s.WriteString("\n")
		for i, f := range v.Untracked {
			prefix := "  "
			if i+v.LenStaged()+v.LenUnstaged() == v.Selected {
				prefix = styles.SelectedStyle.Render("> ")
			}
			s.WriteString(prefix + f.Path + "\n")
		}
	}

	return styles.PanelStyle.Render(s.String())
}

func (v StatusView) LenStaged() int    { return len(v.Staged) }
func (v StatusView) LenUnstaged() int  { return len(v.Unstaged) }
func (v StatusView) LenUntracked() int { return len(v.Untracked) }
func (v StatusView) Total() int        { return v.LenStaged() + v.LenUnstaged() + v.LenUntracked() }

func (v *StatusView) Select(i int) {
	if i >= 0 && i < v.Total() {
		v.Selected = i
	}
}

func (v *StatusView) MoveUp() {
	if v.Selected > 0 {
		v.Selected--
	}
}

func (v *StatusView) MoveDown() {
	if v.Selected < v.Total()-1 {
		v.Selected++
	}
}

func (v *StatusView) SelectedFile() *entities.FileStatus {
	idx := v.Selected
	if idx < v.LenStaged() {
		return &v.Staged[idx]
	}
	idx -= v.LenStaged()
	if idx < v.LenUnstaged() {
		return &v.Unstaged[idx]
	}
	idx -= v.LenUnstaged()
	if idx < v.LenUntracked() {
		return &v.Untracked[idx]
	}
	return nil
}

func truncateStr(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}
