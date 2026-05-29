package views

import (
	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/styles"
)

type DiffView struct {
	Diff *entities.Diff
}

func NewDiffView() DiffView {
	return DiffView{
		Diff: nil,
	}
}

func (v DiffView) Update(diff *entities.Diff) DiffView {
	v.Diff = diff
	return v
}

func (v DiffView) Render() string {
	if v.Diff == nil {
		return "No diff available"
	}

	var s string

	s += styles.TitleStyle.Render("Diff: "+v.Diff.FilePath) + "\n\n"

	for _, hunk := range v.Diff.Hunks {
		for _, line := range hunk.Lines {
			switch line.Type {
			case entities.DiffLineAdded:
				s += styles.DiffAddedStyle.Render("+ " + line.Content)
			case entities.DiffLineDeleted:
				s += styles.DiffRemovedStyle.Render("- " + line.Content)
			case entities.DiffLineHeader:
				s += styles.AccentStyle.Render(line.Content)
			default:
				s += " " + line.Content
			}
			s += "\n"
		}
	}

	return s
}