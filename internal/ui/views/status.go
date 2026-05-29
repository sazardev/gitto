package views

import (
	
	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/styles"
)

type StatusView struct {
	Staged    []entities.FileStatus
	Unstaged  []entities.FileStatus
	Untracked []entities.FileStatus
	Selected  int
	Height    int
}

func NewStatusView() StatusView {
	return StatusView{
		Staged:    []entities.FileStatus{},
		Unstaged:  []entities.FileStatus{},
		Untracked: []entities.FileStatus{},
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

	return v
}

func (v StatusView) Render() string {
	var s string

	s += styles.TitleStyle.Render("Status") + "\n\n"

	if len(v.Staged) > 0 {
		s += styles.StagedStyle.Render("Staged") + "\n"
		for i, f := range v.Staged {
			prefix := "  "
			if i == v.Selected {
				prefix = styles.SelectedStyle.Render(">")
			}
			s += prefix + " " + f.Path + "\n"
		}
		s += "\n"
	}

	if len(v.Unstaged) > 0 {
		s += styles.UnstagedStyle.Render("Unstaged") + "\n"
		for i, f := range v.Unstaged {
			prefix := "  "
			if i+v.LenStaged() == v.Selected {
				prefix = styles.SelectedStyle.Render(">")
			}
			s += prefix + " " + f.Path + "\n"
		}
		s += "\n"
	}

	if len(v.Untracked) > 0 {
		s += styles.UntrackedStyle.Render("Untracked") + "\n"
		for i, f := range v.Untracked {
			prefix := "  "
			if i+v.LenStaged()+v.LenUnstaged() == v.Selected {
				prefix = styles.SelectedStyle.Render(">")
			}
			s += prefix + " " + f.Path + "\n"
		}
	}

	return s
}

func (v StatusView) LenStaged() int   { return len(v.Staged) }
func (v StatusView) LenUnstaged() int { return len(v.Unstaged) }
func (v StatusView) LenUntracked() int { return len(v.Untracked) }
func (v StatusView) Total() int       { return v.LenStaged() + v.LenUnstaged() + v.LenUntracked() }

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