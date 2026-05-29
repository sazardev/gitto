package views

import (
	
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
	var s string

	s += styles.TitleStyle.Render("Log") + "\n\n"

	for i, c := range v.Commits {
		prefix := "  "
		if i == v.Selected {
			prefix = styles.SelectedStyle.Render(">")
		}
		s += prefix + " " + c.ShortHash + " | " + c.Author + "\n"
		s += "    " + c.Message + "\n"
	}

	return s
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