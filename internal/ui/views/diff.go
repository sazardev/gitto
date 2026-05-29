package views

import (
	"fmt"
	"strings"

	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/styles"
)

type DiffView struct {
	Diff           *entities.Diff
	ScrollOffset   int
	SelectedHunk   int
	FilePath       string
	IsStaged       bool
	SelectedLine   int
	LineMode       bool
}

func NewDiffView() DiffView {
	return DiffView{
		Diff:         nil,
		SelectedHunk: 0,
		SelectedLine: 0,
		LineMode:     false,
	}
}

func (v DiffView) Update(diff *entities.Diff) DiffView {
	v.Diff = diff
	v.ScrollOffset = 0
	v.SelectedHunk = 0
	v.SelectedLine = 0
	return v
}

func (v DiffView) SetFile(path string, staged bool) DiffView {
	v.FilePath = path
	v.IsStaged = staged
	return v
}

func (v DiffView) Render(width, height int) string {
	if v.Diff == nil {
		return styles.PanelStyle.Width(width).Height(height).Render(
			styles.PanelTitleStyle.Render("[d] Diff") + "\n\n" +
				styles.DimStyle.Render("  no diff available"),
		)
	}

	stagedLabel := "unstaged"
	if v.IsStaged {
		stagedLabel = "staged"
	}
	title := styles.PanelTitleStyle.Render(fmt.Sprintf("[d] Diff (%s)", stagedLabel))

	hunkCount := len(v.Diff.Hunks)
	if hunkCount == 0 {
		return styles.PanelStyle.Width(width).Height(height).Render(
			title + "\n\n" + styles.DimStyle.Render("  no changes"),
		)
	}

	hunkInfo := styles.DimStyle.Render(fmt.Sprintf("  hunk %d/%d", v.SelectedHunk+1, hunkCount))

	usedLines := 3
	startLine := v.ScrollOffset
	currentLine := 0

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n")
	sb.WriteString(hunkInfo)
	sb.WriteString("\n\n")

	for hIdx, hunk := range v.Diff.Hunks {
		isSelected := hIdx == v.SelectedHunk

		if hIdx > 0 && currentLine < startLine && currentLine+len(hunk.Lines) >= startLine {
			sb.WriteString("\n")
			sb.WriteString(styles.DimStyle.Render("  ···"))
			sb.WriteString("\n")
			usedLines++
		}

		for lIdx, line := range hunk.Lines {
			if currentLine < startLine {
				currentLine++
				continue
			}

			if usedLines >= height-2 {
				sb.WriteString("\n")
				sb.WriteString(styles.DimStyle.Render("  ... scroll with ↑↓"))
				return sb.String()
			}

			content := line.Content
			maxContentWidth := width - 10
			if maxContentWidth > 0 && len(content) > maxContentWidth {
				content = content[:maxContentWidth-3] + "..."
			}

			var prefix string
			isLineSelected := isSelected && v.LineMode && lIdx == v.SelectedLine
			if isLineSelected {
				prefix = styles.SelectedStyle.Render(">> ")
			} else if isSelected {
				prefix = styles.SelectedStyle.Render(">  ")
			} else {
				prefix = "   "
			}

			switch line.Type {
			case entities.DiffLineAdded:
				sb.WriteString(prefix + styles.DiffAddedStyle.Render("+ "+content))
			case entities.DiffLineDeleted:
				sb.WriteString(prefix + styles.DiffRemovedStyle.Render("- "+content))
			case entities.DiffLineHeader:
				sb.WriteString(prefix + styles.DiffHeaderStyle.Render(content))
			default:
				sb.WriteString(prefix + "  " + styles.DimStyle.Render(content))
			}
			sb.WriteString("\n")
			usedLines++
			currentLine++
		}

		if isSelected && hIdx < hunkCount-1 {
			if usedLines < height-2 {
				sb.WriteString(styles.DimStyle.Render("  ────────────────────────────────────"))
				sb.WriteString("\n")
				usedLines++
			}
		}
	}

	return styles.PanelStyle.Width(width).Height(height).Render(sb.String())
}

func (v *DiffView) MoveUp() {
	if v.LineMode && v.Diff != nil && v.SelectedHunk < len(v.Diff.Hunks) {
		if v.SelectedLine > 0 {
			v.SelectedLine--
			return
		}
	}
	if v.ScrollOffset > 0 {
		v.ScrollOffset--
	}
}

func (v *DiffView) MoveDown() {
	if v.LineMode && v.Diff != nil && v.SelectedHunk < len(v.Diff.Hunks) {
		hunk := v.Diff.Hunks[v.SelectedHunk]
		if v.SelectedLine < len(hunk.Lines)-1 {
			v.SelectedLine++
			return
		}
	}
	v.ScrollOffset++
}

func (v *DiffView) MoveHunkUp() {
	if v.SelectedHunk > 0 {
		v.SelectedHunk--
		v.SelectedLine = 0
		v.scrollToSelectedHunk()
	}
}

func (v *DiffView) MoveHunkDown() {
	if v.Diff != nil && v.SelectedHunk < len(v.Diff.Hunks)-1 {
		v.SelectedHunk++
		v.SelectedLine = 0
		v.scrollToSelectedHunk()
	}
}

func (v *DiffView) ToggleLineMode() {
	v.LineMode = !v.LineMode
	v.SelectedLine = 0
}

func (v DiffView) scrollToSelectedHunk() {
	if v.Diff == nil || v.SelectedHunk >= len(v.Diff.Hunks) {
		return
	}

	lineCount := 0
	for i := 0; i < v.SelectedHunk; i++ {
		lineCount += len(v.Diff.Hunks[i].Lines)
	}

	v.ScrollOffset = lineCount
}

func (v DiffView) SelectedHunkIndex() int {
	return v.SelectedHunk
}

func (v DiffView) HasDiff() bool {
	return v.Diff != nil && len(v.Diff.Hunks) > 0
}
