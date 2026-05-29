package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/styles"
)

type LogView struct {
	Commits  []entities.Commit
	Selected int
	Height   int
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

func (v LogView) Render() string {
	var s strings.Builder

	s.WriteString(styles.TitleStyle.Render("Recent Commits"))
	s.WriteString("\n\n")

	if len(v.Commits) == 0 {
		s.WriteString(styles.HelpStyle.Render("No commits found"))
		return s.String()
	}

	for i, c := range v.Commits {
		prefix := "  "
		if i == v.Selected {
			prefix = styles.SelectedStyle.Render("▶ ")
		}
		s.WriteString(prefix)
		s.WriteString(styles.HashStyle.Render(c.ShortHash))
		s.WriteString(" ")
		s.WriteString(styles.DimStyle.Render(c.Author))
		s.WriteString(" ")
		s.WriteString(styles.TimeStyle.Render(formatTime(c.AuthorDate)))
		s.WriteString("\n")

		msg := truncate(c.Message, 60)
		s.WriteString("    ")
		s.WriteString(msg)
		s.WriteString("\n\n")
	}

	s.WriteString(styles.HelpStyle.Render("↑↓ navigate • l toggle • Esc back"))
	return s.String()
}

func (v *LogView) MoveUp() {
	if v.Selected > 0 {
		v.Selected--
	}
}

func (v *LogView) MoveDown() {
	if v.Selected < len(v.Commits)-1 {
		v.Selected++
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

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}
