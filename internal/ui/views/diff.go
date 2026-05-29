package views

import (
	"strings"

	"github.com/sazardev/gitto/internal/styles"
)

type DiffView struct {
	Content  string
	FilePath string
	Lines    []string
}

func NewDiffView() DiffView {
	return DiffView{
		Content:  "",
		FilePath: "",
		Lines:    []string{},
	}
}

func (v DiffView) Update(content, filePath string) DiffView {
	v.Content = content
	v.FilePath = filePath
	v.Lines = strings.Split(content, "\n")
	return v
}

func (v DiffView) Render() string {
	var s string

	s += styles.TitleStyle.Render("Diff: "+v.FilePath) + "\n\n"

	for _, line := range v.Lines {
		switch {
		case strings.HasPrefix(line, "+"):
			s += styles.DiffAddedStyle.Render(line) + "\n"
		case strings.HasPrefix(line, "-"):
			s += styles.DiffRemovedStyle.Render(line) + "\n"
		case strings.HasPrefix(line, "@"):
			s += styles.AccentStyle.Render(line) + "\n"
		default:
			s += line + "\n"
		}
	}

	return s
}