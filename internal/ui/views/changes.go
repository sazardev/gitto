package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/styles"
)

type ChangesView struct {
	Staged         []entities.FileStatus
	Unstaged       []entities.FileStatus
	Untracked      []entities.FileStatus
	Selected       int
	ScrollOffset   int
	StagedCount    int
	UnstagedCount  int
	UntrackedCount int
}

func NewChangesView() ChangesView {
	return ChangesView{
		Staged:    []entities.FileStatus{},
		Unstaged:  []entities.FileStatus{},
		Untracked: []entities.FileStatus{},
		Selected:  0,
	}
}

func (v ChangesView) Update(files []entities.FileStatus) ChangesView {
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

func (v ChangesView) Render(width, height int) string {
	if width < 40 {
		return v.renderCompact(width, height)
	}
	return v.renderFull(width, height)
}

func (v ChangesView) renderFull(width, height int) string {
	title := styles.TitleStyle.Render("changes")

	total := v.Total()
	if total == 0 {
		return title + "\n\n" + styles.DimStyle.Render("  no changes")
	}

	statsLine := fmt.Sprintf("%s staged  %s unstaged  %s untracked",
		styles.StatStagedStyle.Render(fmt.Sprintf("%d", v.StagedCount)),
		styles.StatUnstagedStyle.Render(fmt.Sprintf("%d", v.UnstagedCount)),
		styles.StatUntrackedStyle.Render(fmt.Sprintf("%d", v.UntrackedCount)),
	)

	availableHeight := height - 5
	if availableHeight < 3 {
		availableHeight = 3
	}

	allFiles := v.getAllFiles()
	startIdx := v.ScrollOffset
	endIdx := startIdx + availableHeight
	if endIdx > len(allFiles) {
		endIdx = len(allFiles)
	}

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n")
	sb.WriteString(statsLine)
	sb.WriteString("\n\n")

	for i := startIdx; i < endIdx; i++ {
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
		maxPW := width - 12
		if maxPW > 0 && len(path) > maxPW {
			path = path[:maxPW-3] + "..."
		}

		sb.WriteString(prefix + s.Render(icon) + " " + path + "\n")
	}

	if total > availableHeight {
		sb.WriteString("\n")
		sb.WriteString(styles.DimStyle.Render(fmt.Sprintf("  %d/%d files", endIdx, total)))
	}

	sb.WriteString("\n")
	sb.WriteString(styles.HelpStyle.Render("↑↓ navigate • [s] stage • [u] unstage • [d] diff • esc back"))

	return sb.String()
}

func (v ChangesView) renderCompact(width, height int) string {
	title := styles.TitleStyle.Render("changes")

	s := fmt.Sprintf("%s%s%s",
		styles.StatStagedStyle.Render(fmt.Sprintf("%ds", v.StagedCount)),
		styles.StatUnstagedStyle.Render(fmt.Sprintf(" %du", v.UnstagedCount)),
		styles.StatUntrackedStyle.Render(fmt.Sprintf(" %d?", v.UntrackedCount)),
	)

	return title + "\n\n" + s
}

func (v ChangesView) getAllFiles() []entities.FileStatus {
	var all []entities.FileStatus
	all = append(all, v.Staged...)
	all = append(all, v.Unstaged...)
	all = append(all, v.Untracked...)
	return all
}

func (v *ChangesView) MoveUp() {
	if v.Selected > 0 {
		v.Selected--
		v.adjustScroll()
	}
}

func (v *ChangesView) MoveDown() {
	if v.Selected < v.Total()-1 {
		v.Selected++
		v.adjustScroll()
	}
}

func (v *ChangesView) adjustScroll() {
	if v.Selected < v.ScrollOffset {
		v.ScrollOffset = v.Selected
	}
}

func (v ChangesView) SelectedFile() *entities.FileStatus {
	all := v.getAllFiles()
	if v.Selected >= 0 && v.Selected < len(all) {
		return &all[v.Selected]
	}
	return nil
}

func (v ChangesView) Total() int {
	return v.LenStaged() + v.LenUnstaged() + v.LenUntracked()
}

func (v ChangesView) LenStaged() int    { return len(v.Staged) }
func (v ChangesView) LenUnstaged() int  { return len(v.Unstaged) }
func (v ChangesView) LenUntracked() int { return len(v.Untracked) }
