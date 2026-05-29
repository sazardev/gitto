package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/styles"
)

type LogView struct {
	Commits      []entities.Commit
	Selected     int
	ScrollOffset int
}

func NewLogView() LogView {
	return LogView{
		Commits:  []entities.Commit{},
		Selected: 0,
	}
}

func (v LogView) Update(commits []entities.Commit) LogView {
	v.Commits = commits
	return v
}

func (v LogView) Render(width, height int) string {
	if width < 40 {
		return v.renderCompact(width, height)
	}
	return v.renderFull(width, height)
}

func (v LogView) renderFull(width, height int) string {
	title := styles.PanelTitleStyle.Render("[4] Commits")

	if len(v.Commits) == 0 {
		return title + "\n\n" + styles.HelpStyle.Render("  no commits found")
	}

	showTime := width >= 60
	showAuthor := width >= 50

	availableHeight := height - 4
	if availableHeight < 3 {
		availableHeight = 3
	}

	startIdx := v.ScrollOffset
	endIdx := startIdx + availableHeight/2
	if endIdx > len(v.Commits) {
		endIdx = len(v.Commits)
	}

	msgWidth := width - 20
	if msgWidth < 15 {
		msgWidth = 15
	}
	if width < 50 {
		msgWidth = width - 14
	}

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n\n")

	for i := startIdx; i < endIdx; i++ {
		c := v.Commits[i]
		prefix := "  "
		if i == v.Selected {
			prefix = styles.SelectedStyle.Render("▶ ")
		}
		sb.WriteString(prefix)
		sb.WriteString(styles.HashStyle.Render(c.ShortHash))
		sb.WriteString(" ")

		if showAuthor {
			authorWidth := 12
			if width < 70 {
				authorWidth = 8
			}
			sb.WriteString(styles.DimStyle.Render(truncateStr(c.Author, authorWidth)))
			sb.WriteString(" ")
		}

		if showTime {
			sb.WriteString(styles.TimeStyle.Render(formatTime(c.AuthorDate)))
			sb.WriteString(" ")
		}

		sb.WriteString("\n")

		msg := truncateStr(c.Message, msgWidth)
		sb.WriteString("    ")
		sb.WriteString(msg)
		sb.WriteString("\n\n")
	}

	if len(v.Commits) > availableHeight/2 {
		sb.WriteString(styles.DimStyle.Render(fmt.Sprintf("  %d/%d commits", endIdx, len(v.Commits))))
		sb.WriteString("\n")
	}

	sb.WriteString(styles.HelpStyle.Render("↑↓ navigate • esc back"))

	return styles.PanelStyle.Width(width).Height(height).Render(sb.String())
}

func (v LogView) renderCompact(width, height int) string {
	title := styles.TitleStyle.Render("history")

	if len(v.Commits) == 0 {
		return title + "\n" + styles.DimStyle.Render("  none")
	}

	maxCommits := height / 3
	if maxCommits < 1 {
		maxCommits = 1
	}
	if maxCommits > len(v.Commits) {
		maxCommits = len(v.Commits)
	}

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n\n")

	for i := 0; i < maxCommits; i++ {
		c := v.Commits[i]
		sb.WriteString(styles.HashStyle.Render(c.ShortHash))
		sb.WriteString(" ")
		sb.WriteString(truncateStr(c.Message, width-10))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (v *LogView) MoveUp() {
	if v.Selected > 0 {
		v.Selected--
		v.adjustScroll()
	}
}

func (v *LogView) MoveDown() {
	if v.Selected < len(v.Commits)-1 {
		v.Selected++
		v.adjustScroll()
	}
}

func (v *LogView) adjustScroll() {
	if v.Selected < v.ScrollOffset {
		v.ScrollOffset = v.Selected
	}
}

func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else {
		return t.Format("Jan 02")
	}
}
