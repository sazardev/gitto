# Plan: Diff View Real

## Problema Actual
DiffView muestra el contenido del archivo en HEAD, no las diferencias reales entre:
- staged vs HEAD (para archivos en Staged)
- worktree vs staged (para archivos en Unstaged)

## Flujo Esperado
1. Usuario navega a un archivo modificado
2. Presiona `d` o `<Enter>`
3. Se abre DiffView modal
4. Muestra diff con formato:
   - Líneas añadidas: verde (`+`)
   - Líneas eliminadas: rojo (`-`)
   - Header: `diff --git a/file b/file`
5. `<Esc>` o `q` cierra el modal

## Implementación

### 1. Gogit adapter - Add Diff method
**File:** `internal/adapters/gogit/diff.go` (crear)

```go
package gogit

import (
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"
    "github.com/go-git/go-git/v5/utils/filemap"
)

type Diff struct {
    File  string
    Lines []DiffLine
}

type DiffLine struct {
    Type    string // "+", "-", " ", ""
    Content string
}

func (g *GoGitAdapter) GetDiff(filePath string, staged bool) (*Diff, error) {
    wt, err := g.repo.Worktree()
    if err != nil {
        return nil, err
    }

    var fromCommit *object.Commit
    var toTree *object.Tree

    if staged {
        // Staged: compare HEAD vs staged (index)
        fromCommit, err = g.repo.CommitObject(g.repo.Head().Hash())
        if err != nil {
            return nil, err
        }
        fromTree, err := fromCommit.Trie()
        if err != nil {
            return nil, err
        }
        
        // Get index/staged tree
        idx, err := wt.Staging()
        if err != nil {
            return nil, err
        }
        toTree, err = idx.Trie()
        if err != nil {
            return nil, err
        }
    } else {
        // Unstaged: compare staged vs worktree
        fromTree, err = wt.Staging().Trie()
        if err != nil {
            return nil, err
        }
        toTree, err = wt.Trie()
        if err != nil {
            return nil, err
        }
    }

    // Get file content changes
    diff, err := fromTree.Diff(toTree)
    if err != nil {
        return nil, err
    }

    // Find the specific file diff
    for _, d := range diff {
        if d.Files[0].Name == filePath || d.Files[1].Name == filePath {
            return convertFileDiffToDiff(d, filePath)
        }
    }

    // Fallback: file not in diff, show current state
    return &Diff{File: filePath, Lines: []DiffLine{
        {Type: " ", Content: "(file not in diff)"},
    }}, nil
}

func convertFileDiffToDiff(d object.FileDiff, filePath string) (*Diff, error) {
    result := &Diff{
        File:  filePath,
        Lines: []DiffLine{},
    }

    // Parse the patch
    patch, err := d.Patch()
    if err != nil {
        return nil, err
    }

    for _, filePatch := range patch.FilePatches {
        for _, hunk := range filePatch.Hunks {
            for _, ln := range hunk.Lines {
                diffLine := DiffLine{
                    Content: ln.Content,
                }
                switch ln.Type {
                case 1: // Add
                    diffLine.Type = "+"
                case 2: // Delete
                    diffLine.Type = "-"
                default:
                    diffLine.Type = " "
                }
                result.Lines = append(result.Lines, diffLine)
            }
        }
    }

    return result, nil
}
```

### 2. GitProvider interface - add GetDiff
**File:** `internal/ports/git.go`

```go
type GitProvider interface {
    // ... existing methods
    GetDiff(path string, staged bool) (*core.Diff, error)
}
```

### 3. Add Diff to core entities
**File:** `internal/core/entities/diff.go`

```go
package entities

type Diff struct {
    File  string
    Lines []DiffLine
}

type DiffLine struct {
    Type    string // "+", "-", " "
    Content string
}
```

### 4. MainModel - handle open/close DiffView
**File:** `internal/ui/model.go`

```go
type ViewState int

const (
    StatusView ViewState = iota
    DiffView
    LogView
    CommitView
)

type MainModel struct {
    // ... existing fields
    viewState ViewState
    diffView  DiffView
    diffData  *core.Diff
    cursor    int
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    // ... existing cases
    
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keyBindings.Diff):
            if m.viewState == StatusView {
                m.viewState = DiffView
                file := m.status.Files[m.cursor]
                staged := file.Status == entities.StatusStaged
                return m, func() tea.Msg {
                    diff, err := m.git.GetDiff(file.Path, staged)
                    if err != nil {
                        return DiffError{err}
                    }
                    return DiffLoaded{diff}
                }
            }
        case key.Matches(msg, key.NewBinding(tea.KeyEsc, tea.KeyCtrlC)):
            if m.viewState == DiffView {
                m.viewState = StatusView
                return m, nil
            }
        }
    }
}

func (m *MainModel) View() string {
    switch m.viewState {
    case DiffView:
        return m.diffView.View(m.diffData)
    default:
        return m.statusView.View(m.status, m.cursor)
    }
}
```

### 5. DiffView component
**File:** `internal/ui/views/diff.go` (crear)

```go
package views

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/gitto/gitto/internal/core/entities"
    "github.com/gitto/gitto/internal/styles"
)

type DiffView struct{}

func NewDiffView() DiffView {
    return DiffView{}
}

func (v DiffView) View(diff *entities.Diff) string {
    if diff == nil {
        return "No diff available"
    }

    var s strings.Builder
    s.WriteString(styles.BoldStyle.Render("diff --git a/" + diff.File + " b/" + diff.File))
    s.WriteString("\n\n")

    for _, line := range diff.Lines {
        switch line.Type {
        case "+":
            s.WriteString(styles.GreenStyle.Render("+ " + line.Content))
        case "-":
            s.WriteString(styles.RedStyle.Render("- " + line.Content))
        default:
            s.WriteString(line.Content)
        }
        s.WriteString("\n")
    }

    s.WriteString("\n")
    s.WriteString(styles.DimStyle.Render("Press Esc to close"))

    return s.String()
}
```

### 6. Add styles for diff
**File:** `internal/styles/styles.go`

```go
var GreenStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#2E7D32"))

var RedStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#C62828"))
```

## Verificación
- [ ] Staged file → `d` muestra diff staged vs HEAD
- [ ] Unstaged file → `d` muestra diff worktree vs staged
- [ ] `+` líneas en verde
- [ ] `-` líneas en rojo
- [ ] `<Esc>` cierra el modal
