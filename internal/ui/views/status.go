package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/styles"
)

type StatusView struct {
	Staged         []entities.FileStatus
	Unstaged       []entities.FileStatus
	Untracked      []entities.FileStatus
	Branches       []entities.Branch
	Commits        []entities.Commit
	Selected       int
	ScrollOffset   int
	StagedCount    int
	UnstagedCount  int
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

func (v StatusView) Render(width, height int) string {
	if width < 50 {
		return v.renderNarrow(width, height)
	}
	return v.renderWide(width, height)
}

func (v StatusView) renderWide(width, height int) string {
	gap := 2
	leftWidth := (width - gap) * 38 / 100
	rightWidth := width - leftWidth - gap

	leftHeight := height
	rightHeight := height

	leftCol := lipgloss.JoinVertical(lipgloss.Top,
		v.renderBranchesPanel(leftWidth, leftHeight*45/100),
		"\n",
		v.renderChangesSummary(leftWidth, leftHeight*35/100),
	)

	rightCol := v.renderCommitsPanel(rightWidth, rightHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "  ", rightCol)
}

func (v StatusView) renderNarrow(width, height int) string {
	branchH := height * 25 / 100
	commitsH := height * 40 / 100
	changesH := height - branchH - commitsH - 2
	if changesH < 3 { changesH = 3 }

	var sb strings.Builder
	sb.WriteString(v.renderBranchesPanel(width-2, branchH))
	sb.WriteString("\n")
	sb.WriteString(v.renderCommitsPanel(width-2, commitsH))
	sb.WriteString("\n")
	sb.WriteString(v.renderChangesSummary(width-2, changesH))
	return sb.String()
}

func (v StatusView) renderBranchesPanel(panelWidth, maxHeight int) string {
	title := styles.PanelTitleStyle.Render("branches")

	if len(v.Branches) == 0 {
		content := title + "\n" + styles.DimStyle.Render("  no branches")
		return styles.PanelStyle.Width(panelWidth).Render(content)
	}

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n")

	visible := maxHeight - 3
	if visible < 1 { visible = 1 }
	if visible > len(v.Branches) { visible = len(v.Branches) }

	for i := 0; i < visible; i++ {
		b := v.Branches[i]
		prefix := "  "
		var s lipgloss.Style

		if b.IsHead {
			prefix = "> "
			s = styles.BranchActiveStyle
		} else if b.IsRemote {
			s = styles.BranchRemoteStyle
		} else {
			s = styles.BranchItemStyle
		}

		name := b.Name
		maxW := panelWidth - 10
		if maxW > 0 && len(name) > maxW {
			name = name[:maxW-3] + "..."
		}

		sb.WriteString(prefix + s.Render(name))
		sb.WriteString("\n")
	}

	if len(v.Branches) > visible {
		sb.WriteString(styles.DimStyle.Render(fmt.Sprintf("  %d/%d", visible, len(v.Branches))))
	}

	return styles.PanelStyle.Width(panelWidth).Render(sb.String())
}

func (v StatusView) renderCommitsPanel(panelWidth, maxHeight int) string {
	title := styles.PanelTitleStyle.Render("recent commits")

	if len(v.Commits) == 0 {
		content := title + "\n" + styles.DimStyle.Render("  no commits found")
		return styles.PanelStyle.Width(panelWidth).Render(content)
	}

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n")

	visible := maxHeight - 3
	if visible < 1 { visible = 1 }
	if visible > len(v.Commits) { visible = len(v.Commits) }

	hashW := 8
	timeW := 8
	authorW := 10

	if panelWidth < 55 {
		authorW = 0
	}
	if panelWidth < 45 {
		timeW = 0
	}

	msgW := panelWidth - hashW - timeW - authorW - 14
	if msgW < 10 { msgW = 10 }

	for i := 0; i < visible; i++ {
		c := v.Commits[i]
		sb.WriteString(styles.CommitHashStyle.Render(truncateStr(c.ShortHash, hashW)))
		sb.WriteString(" ")

		if authorW > 0 {
			sb.WriteString(styles.CommitAuthorStyle.Render(truncateStr(c.Author, authorW)))
			sb.WriteString(" ")
		}

		if timeW > 0 {
			sb.WriteString(styles.CommitTimeStyle.Render(truncateStr(formatTime(c.AuthorDate), timeW)))
			sb.WriteString(" ")
		}

		sb.WriteString(styles.CommitMsgStyle.Render(truncateStr(c.Message, msgW)))
		sb.WriteString("\n")
	}

	if len(v.Commits) > visible {
		sb.WriteString(styles.DimStyle.Render(fmt.Sprintf("  %d/%d commits", visible, len(v.Commits))))
	}

	return styles.PanelStyle.Width(panelWidth).Render(sb.String())
}

func (v StatusView) renderChangesSummary(panelWidth, maxHeight int) string {
	title := styles.PanelTitleStyle.Render("changes")

	total := v.Total()
	if total == 0 {
		content := title + "\n" + styles.DimStyle.Render("  no changes")
		return styles.PanelStyle.Width(panelWidth).Render(content)
	}

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n")

	stats := fmt.Sprintf("%s staged  %s unstaged  %s untracked",
		styles.StatStagedStyle.Render(fmt.Sprintf("%d", v.StagedCount)),
		styles.StatUnstagedStyle.Render(fmt.Sprintf("%d", v.UnstagedCount)),
		styles.StatUntrackedStyle.Render(fmt.Sprintf("%d", v.UntrackedCount)),
	)
	sb.WriteString(stats)

	visible := maxHeight - 4
	if visible > 0 && total > 0 {
		sb.WriteString("\n\n")
		allFiles := v.getAllFiles()
		if visible > len(allFiles) {
			visible = len(allFiles)
		}
		for i := 0; i < visible; i++ {
			f := allFiles[i]
			prefix := "  "
			if i == v.Selected {
				prefix = styles.SelectedStyle.Render("> ")
			}

			var icon string
			var s lipgloss.Style
			if f.IsStaged {
				icon, s = "+", styles.StagedStyle
			} else if f.IsUntracked {
				icon, s = "?", styles.UntrackedStyle
			} else {
				icon, s = "~", styles.UnstagedStyle
			}

			path := f.Path
			maxPW := panelWidth - 12
			if maxPW > 0 && len(path) > maxPW {
				path = path[:maxPW-3] + "..."
			}

			sb.WriteString(prefix + s.Render(icon) + " " + path + "\n")
		}
	}

	return styles.PanelStyle.Width(panelWidth).Render(sb.String())
}

func (v StatusView) getAllFiles() []entities.FileStatus {
	var all []entities.FileStatus
	all = append(all, v.Staged...)
	all = append(all, v.Unstaged...)
	all = append(all, v.Untracked...)
	return all
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
	if max <= 0 {
		return ""
	}
	if len(s) > max {
		if max <= 3 {
			return s[:max]
		}
		return s[:max-3] + "..."
	}
	return s
}
