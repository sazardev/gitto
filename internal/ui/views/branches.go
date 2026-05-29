package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/styles"
)

type BranchesView struct {
	Branches     []entities.Branch
	Selected     int
	ScrollOffset int
}

func NewBranchesView() BranchesView {
	return BranchesView{
		Branches: []entities.Branch{},
		Selected: 0,
	}
}

func (v BranchesView) Update(branches []entities.Branch) BranchesView {
	v.Branches = branches
	return v
}

func (v BranchesView) Render(width, height int) string {
	if width < 40 {
		return v.renderCompact(width, height)
	}
	return v.renderFull(width, height)
}

func (v BranchesView) renderFull(width, height int) string {
	title := styles.PanelTitleStyle.Render("[2] Branches")

	if len(v.Branches) == 0 {
		return title + "\n\n" + styles.DimStyle.Render("  no branches found")
	}

	localCount := 0
	remoteCount := 0
	for _, b := range v.Branches {
		if b.IsRemote {
			remoteCount++
		} else {
			localCount++
		}
	}

	statsLine := fmt.Sprintf("%s local  %s remote",
		styles.StatStagedStyle.Render(fmt.Sprintf("%d", localCount)),
		styles.StatUntrackedStyle.Render(fmt.Sprintf("%d", remoteCount)),
	)

	availableHeight := height - 5
	if availableHeight < 3 {
		availableHeight = 3
	}

	startIdx := v.ScrollOffset
	endIdx := startIdx + availableHeight
	if endIdx > len(v.Branches) {
		endIdx = len(v.Branches)
	}

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n")
	sb.WriteString(statsLine)
	sb.WriteString("\n\n")

	for i := startIdx; i < endIdx; i++ {
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

		if i == v.Selected {
			prefix = styles.SelectedStyle.Render("> ")
		}

		name := b.Name
		maxW := width - 10
		if maxW > 0 && len(name) > maxW {
			name = name[:maxW-3] + "..."
		}

		sb.WriteString(prefix + s.Render(name))
		sb.WriteString("\n")
	}

	if len(v.Branches) > availableHeight {
		sb.WriteString("\n")
		sb.WriteString(styles.DimStyle.Render(fmt.Sprintf("  %d/%d branches", endIdx, len(v.Branches))))
	}

	sb.WriteString("\n")
	sb.WriteString(styles.HelpStyle.Render("↑↓ navigate • enter checkout • esc back"))

	return styles.PanelStyle.Width(width).Height(height).Render(sb.String())
}

func (v BranchesView) renderCompact(width, height int) string {
	title := styles.TitleStyle.Render("branches")

	if len(v.Branches) == 0 {
		return title + "\n" + styles.DimStyle.Render("  none")
	}

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n")

	for _, b := range v.Branches {
		if b.IsHead {
			sb.WriteString(styles.BranchActiveStyle.Render("> " + b.Name))
			break
		}
	}

	return sb.String()
}

func (v *BranchesView) MoveUp() {
	if v.Selected > 0 {
		v.Selected--
		v.adjustScroll()
	}
}

func (v *BranchesView) MoveDown() {
	if v.Selected < len(v.Branches)-1 {
		v.Selected++
		v.adjustScroll()
	}
}

func (v *BranchesView) adjustScroll() {
	if v.Selected < v.ScrollOffset {
		v.ScrollOffset = v.Selected
	}
}

func (v BranchesView) SelectedBranch() *entities.Branch {
	if v.Selected >= 0 && v.Selected < len(v.Branches) {
		return &v.Branches[v.Selected]
	}
	return nil
}

func (v BranchesView) Total() int {
	return len(v.Branches)
}
