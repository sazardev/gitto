# Plan: Log View Funcional

## Problema Actual
LogView existe pero:
- No se actualiza cuando se abre
- No muestra commits reales
- El cursor no funciona para navegar

## Flujo Esperado
1. Usuario presiona `l` → LogView se muestra con commits recientes
2. Commits cargan de forma async (spinner inicial si es lento)
3. Usuario puede navegar con `j`/`k` o `↑`/`↓`
4. `<Enter>` sobre un commit → podría mostrar detalles (opcional para MVP)
5. `<Esc>` o `l` de nuevo → vuelve a StatusView

## Implementación

### 1. Gogit adapter - Log method
**File:** `internal/adapters/gogit/log.go` (crear)

```go
package gogit

import (
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/object"
)

const DefaultLogCount = 50

func (g *GoGitAdapter) Log(count int) ([]*Commit, error) {
    if count <= 0 {
        count = DefaultLogCount
    }

    ref, err := g.repo.Head()
    if err != nil {
        return nil, err
    }

    commitIter, err := g.repo.Log(&git.LogOptions{
        From: ref.Hash(),
    })
    if err != nil {
        return nil, err
    }

    var commits []*Commit
    for i := 0; i < count; i++ {
        commit, err := commitIter.Next()
        if err != nil {
            break // EOF
        }
        commits = append(commits, convertToCommit(commit))
    }

    return commits, nil
}

func convertToCommit(c *object.Commit) *Commit {
    return &Commit{
        Hash:      c.Hash.String(),
        ShortHash: c.Hash.String()[:7],
        Message:   c.Message,
        Author:    c.Author.Name,
        Email:     c.Author.Email,
        Date:      c.Author.When,
        Parents:   c.ParentHashes,
    }
}
```

### 2. Update GitProvider interface
**File:** `internal/ports/git.go`

```go
type GitProvider interface {
    Status() (*RepositoryStatus, error)
    StageFile(path string) error
    UnstageFile(path string) error
    Commit(message string) (string, error)
    Log(count int) ([]*Commit, error)
    GetDiff(path string, staged bool) (*Diff, error)
    Push() error
    Pull() error
    Branches() ([]*Branch, error)
}
```

### 3. MainModel - add log-related fields and state
**File:** `internal/ui/model.go`

```go
type MainModel struct {
    // ... existing fields
    viewState   ViewState
    log         []*Commit
    logCursor   int
    statusView  StatusView
    logView     LogView
}

func NewMainModel(git GitProvider, config ConfigProvider) MainModel {
    return MainModel{
        viewState:  StatusView,
        logCursor:  0,
        statusView: NewStatusView(),
        logView:    NewLogView(),
    }
}
```

### 4. MainModel.Update - handle log key and LogUpdated message
**File:** `internal/ui/model.go`

```go
case tea.KeyMsg:
    switch {
    case key.Matches(msg, m.keyBindings.Log):
        if m.viewState == StatusView {
            m.viewState = LogView
            m.logCursor = 0
            return m, func() tea.Msg {
                log, err := m.git.Log(50)
                if err != nil {
                    return LogError{err}
                }
                return LogLoaded{log}
            }
        } else if m.viewState == LogView {
            m.viewState = StatusView
            return m, nil
        }
    case key.Matches(msg, key.NewBinding(tea.KeyUp)):
        if m.viewState == LogView && m.logCursor > 0 {
            m.logCursor--
            return m, nil
        }
    case key.Matches(msg, key.NewBinding(tea.KeyDown)):
        if m.viewState == LogView && m.logCursor < len(m.log)-1 {
            m.logCursor++
            return m, nil
        }
    case key.Matches(msg, key.NewBinding(tea.KeyEsc, tea.KeyCtrlC)):
        if m.viewState == LogView {
            m.viewState = StatusView
            return m, nil
        }
    }

case LogLoaded:
    m.log = msg.Commits
    m.logCursor = 0
    return m, nil

case LogError:
    m.lastError = msg.error.Error()
    return m, nil
```

### 5. Message types
**File:** `internal/ui/messages.go`

```go
type LogLoaded struct {
    Commits []*Commit
}

type LogError struct {
    error
}
```

### 6. LogView component
**File:** `internal/ui/views/log.go`

```go
package views

import (
    "fmt"
    "github.com/charmbracelet/lipgloss"
    "github.com/gitto/gitto/internal/core/entities"
    "github.com/gitto/gitto/internal/styles"
)

type LogView struct{}

func NewLogView() LogView {
    return LogView{}
}

func (v LogView) View(commits []*Commit, cursor int) string {
    if len(commits) == 0 {
        return "No commits found"
    }

    var s strings.Builder
    s.WriteString(styles.BoldStyle.Render("Recent Commits"))
    s.WriteString("\n\n")

    for i, commit := range commits {
        prefix := "  "
        if i == cursor {
            prefix = styles.SelectedStyle.Render("▶ ")
        }
        
        s.WriteString(prefix)
        s.WriteString(styles.HashStyle.Render(commit.ShortHash))
        s.WriteString(" ")
        s.WriteString(styles.DimStyle.Render(commit.Author))
        s.WriteString(" ")
        s.WriteString(styles.TimeStyle.Render(formatTime(commit.Date)))
        s.WriteString("\n")
        
        // Wrap message to fit width
        msg := truncate(commit.Message, 60)
        s.WriteString("    ")
        s.WriteString(msg)
        s.WriteString("\n\n")
    }

    s.WriteString(styles.DimStyle.Render("↑↓ navigate • l toggle • Esc back"))
    return s.String()
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
```

### 7. Add styles
**File:** `internal/styles/styles.go`

```go
var HashStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FFB300"))

var TimeStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#9E9E9E"))
```

## Verificación
- [ ] `l` muestra lista de commits
- [ ] Commits muestran hash corto, autor, tiempo relativo, mensaje
- [ ] `↑`/`↓` navegan por los commits
- [ ] Cursor visual indica posición actual
- [ ] `<Esc>` vuelve a StatusView
- [ ] Nuevos commits tras commit/pull aparecen al volver a Log
