package views

import (
	"strings"

	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/styles"
)

type DiffView struct {
	Diff         *entities.Diff
	ScrollOffset int
}

func NewDiffView() DiffView {
	return DiffView{
		Diff: nil,
	}
}

func (v DiffView) Update(diff *entities.Diff) DiffView {
	v.Diff = diff
	v.ScrollOffset = 0
	return v
}

func (v DiffView) Render(width, height int) string {
	if v.Diff == nil {
		return styles.DimStyle.Render("  no diff available")
	}

	title := styles.TitleStyle.Render("diff: " + v.Diff.FilePath)

	usedLines := 2
	startLine := v.ScrollOffset
	currentLine := 0

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n\n")

	for _, hunk := range v.Diff.Hunks {
		for _, line := range hunk.Lines {
			if currentLine < startLine {
				currentLine++
				continue
			}

			if usedLines >= height {
				sb.WriteString("\n")
				sb.WriteString(styles.DimStyle.Render("  ... scroll with ↑↓"))
				return sb.String()
			}

			content := line.Content
			maxContentWidth := width - 4
			if maxContentWidth > 0 && len(content) > maxContentWidth {
				content = content[:maxContentWidth-3] + "..."
			}

			switch line.Type {
			case entities.DiffLineAdded:
				sb.WriteString(styles.DiffAddedStyle.Render("+ " + content))
			case entities.DiffLineDeleted:
				sb.WriteString(styles.DiffRemovedStyle.Render("- " + content))
			case entities.DiffLineHeader:
				sb.WriteString(styles.DiffHeaderStyle.Render(content))
			default:
				sb.WriteString(" " + content)
			}
			sb.WriteString("\n")
			usedLines++
			currentLine++
		}
	}

	return sb.String()
}

func (v *DiffView) MoveUp() {
	if v.ScrollOffset > 0 {
		v.ScrollOffset--
	}
}

func (v *DiffView) MoveDown() {
	v.ScrollOffset++
}
