# Plan: Stage/Unstage con Refresh

## Problema Actual
Stage/Unstage funciona internamente pero NO re-renderiza la vista después de la operación. El usuario no ve el cambio de estado de los archivos.

## Flujo Esperado
1. Usuario selecciona archivo + presiona `s` (stage) o `u` (unstage)
2. Se ejecuta la operación via gogit adapter
3. Se actualiza el modelo internamente
4. Bubble Tea re-renderiza automáticamente via `m.ReturnCmd()` → el ciclo de Update/View se reanuda con el modelo actualizado

## Implementación

### 1. Modify gogit adapter - StageFile/UnstageFile
**File:** `internal/adapters/gogit/worktree.go`

```go
func (g *GoGitAdapter) StageFile(path string) error {
    wt, err := g.repo.Worktree()
    if err != nil {
        return err
    }
    return wt.Add(path)
}

func (g *GoGitAdapter) UnstageFile(path string) error {
    wt, err := g.repo.Worktree()
    if err != nil {
        return err
    }
    // git reset HEAD -- <path>
    return wt.Reset(&git.Reset{Commit: g.repo.Head().Hash(), Path: path})
}
```

### 2. Update MainModel.Update - handle stage/unstage keys
**File:** `internal/ui/model.go`

```go
case tea.KeyMsg:
    switch {
    case key.Matches(msg, m.keyBindings.Stage):
        cmd := m.handleStageFile()
        if cmd != nil {
            return m, cmd
        }
    case key.Matches(msg, m.keyBindings.Unstage):
        cmd := m.handleUnstageFile()
        if cmd != nil {
            return m, cmd
        }
    }
```

### 3. Add handler methods
**File:** `internal/ui/model.go`

```go
func (m *MainModel) handleStageFile() tea.Cmd {
    if m.cursor >= len(m.status.Files) {
        return nil
    }
    file := m.status.Files[m.cursor]
    
    return func() tea.Msg {
        err := m.git.StageFile(file.Path)
        if err != nil {
            return StatusUpdateError{err}
        }
        // Re-fetch status to get updated state
        status, err := m.git.Status()
        if err != nil {
            return StatusUpdateError{err}
        }
        return StatusUpdated{status}
    }
}

func (m *MainModel) handleUnstageFile() tea.Cmd {
    // similar pattern
}
```

### 4. Handle StatusUpdated message in Update
**File:** `internal/ui/model.go`

```go
case StatusUpdated:
    m.status = msg.Status
    m.cursor = 0 // reset cursor after refresh
    return m, nil
```

### 5. Message types
**File:** `internal/ui/messages.go` (crear si no existe)

```go
type StatusUpdated struct {
    Status *core.RepositoryStatus
}

type StatusUpdateError struct {
    error
}
```

## Verificación
- [ ] Stage un archivo → aparece en Staged
- [ ] Unstage un archivo → aparece en Unstaged
- [ ] Múltiples operaciones seguidas funcionan sin crash
