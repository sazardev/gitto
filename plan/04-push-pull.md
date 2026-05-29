# Plan: Push/Pull con Spinner y Feedback

## Problema Actual
Push/Pull existen en el adapter pero:
- No hay feedback visual durante la operación (spinner)
- No hay mensaje de éxito/error al terminar
- No se refresh automatic del log después de pull (nuevos commits)

## Flujo Esperado
1. Usuario presiona `P` (push) o `p` (pull)
2. Spinner aparece con mensaje "Pushing..." o "Pulling..."
3. UI queda bloqueada para input (operación síncrona visual)
4. On success: mensaje de éxito, refresh de status/log
5. On error: mensaje de error descriptivo (no solo "error")

## Implementación

### 1. Message types
**File:** `internal/ui/messages.go`

```go
type PushStarted struct{}
type PushSuccess struct{}
type PushError struct{ error }

type PullStarted struct{}
type PullSuccess struct{}
type PullError struct{ error }
```

### 2. MainModel - add push/pull keybindings
**File:** `internal/ui/model.go`

```go
type KeyBindings struct {
    Stage   key.Binding
    Unstage key.Binding
    Commit  key.Binding
    Log     key.Binding
    Diff    key.Binding
    Push    key.Binding
    Pull    key.Binding
    Quit    key.Binding
}

func DefaultKeyBindings() KeyBindings {
    return KeyBindings{
        Stage:   key.NewBinding(tea.KeyRunes('s'), key.WithHelp("s", "stage")),
        Unstage: key.NewBinding(tea.KeyRunes('u'), key.WithHelp("u", "unstage")),
        Commit:  key.NewBinding(tea.KeyRunes('c'), key.WithHelp("c", "commit")),
        Log:     key.NewBinding(tea.KeyRunes('l'), key.WithHelp("l", "log")),
        Diff:    key.NewBinding(tea.KeyRunes('d'), key.WithHelp("d", "diff")),
        Push:    key.NewBinding(tea.KeyRunes('P'), key.WithHelp("P", "push")),
        Pull:    key.NewBinding(tea.KeyRunes('p'), key.WithHelp("p", "pull")),
        Quit:    key.NewBinding(tea.KeyCtrlC, key.WithHelp("q", "quit")),
    }
}
```

### 3. MainModel.Update - handle push/pull keys
**File:** `internal/ui/model.go`

```go
case tea.KeyMsg:
    switch {
    case key.Matches(msg, m.keyBindings.Push):
        return m, m.handlePush()
    case key.Matches(msg, m.keyBindings.Pull):
        return m, m.handlePull()
    }
```

### 4. Handler methods
**File:** `internal/ui/model.go`

```go
func (m *MainModel) handlePush() tea.Cmd {
    return func() tea.Msg {
        m.mu.Lock()
        m.isLoading = true
        m.loadingMessage = "Pushing..."
        m.mu.Unlock()

        err := m.git.Push()
        if err != nil {
            return PushError{err}
        }
        return PushSuccess{}
    }
}

func (m *MainModel) handlePull() tea.Cmd {
    return func() tea.Msg {
        m.mu.Lock()
        m.isLoading = true
        m.loadingMessage = "Pulling..."
        m.mu.Unlock()

        err := m.git.Pull()
        if err != nil {
            return PullError{err}
        }
        return PullSuccess{}
    }
}
```

### 5. MainModel.Update - handle results
**File:** `internal/ui/model.go`

```go
case PushStarted, PullStarted:
    m.isLoading = true
    m.loadingMessage = extractLoadingMessage(msg)
    return m, nil

case PushSuccess:
    m.isLoading = false
    m.lastMessage = "Push successful"
    return m, RefreshStatusAndLog

case PullSuccess:
    m.isLoading = false
    m.lastMessage = "Pull successful"
    return m, RefreshStatusAndLog

case PushError, PullError:
    m.isLoading = false
    m.lastError = extractError(msg)
    return m, nil
```

### 6. Combined refresh command
**File:** `internal/ui/messages.go`

```go
type RefreshStatusAndLog struct{}

func (m *MainModel) handleRefreshStatusAndLog() tea.Cmd {
    return func() tea.Msg {
        status, err := m.git.Status()
        if err != nil {
            return StatusError{err}
        }
        log, err := m.git.Log(50)
        if err != nil {
            return LogError{err}
        }
        return StatusAndLogRefreshed{status, log}
    }
}
```

### 7. Update View to show spinner + messages
**File:** `internal/ui/model.go`

```go
func (m MainModel) View() string {
    var s strings.Builder

    // Loading state
    if m.isLoading {
        s.WriteString(m.spinner.View())
        s.WriteString(" ")
        s.WriteString(m.loadingMessage)
        s.WriteString("\n\n")
    }

    // Last message (success)
    if m.lastMessage != "" {
        s.WriteString(styles.SuccessStyle.Render("✓ " + m.lastMessage))
        s.WriteString("\n")
    }

    // Last error
    if m.lastError != "" {
        s.WriteString(styles.ErrorStyle.Render("✗ " + m.lastError))
        s.WriteString("\n")
    }

    // Current view
    switch m.viewState {
    case StatusView:
        s.WriteString(m.statusView.View(m.status, m.cursor))
    case DiffView:
        s.WriteString(m.diffView.View(m.diffData))
    case LogView:
        s.WriteString(m.logView.View(m.log, m.cursor))
    }

    return s.String()
}
```

### 8. Gogit adapter - Push/Pull need auth context
**File:** `internal/adapters/gogit/remote.go`

Push/Pull actuales usan `&git.PushOptions{}` sin auth. Para repos remotos privados se necesita:

```go
func (g *GoGitAdapter) Push() error {
    err := g.repo.Push(&git.PushOptions{
        Auth: g.auth, // nil for public repos works fine
    })
    return err
}

func (g *GoGitAdapter) Pull() error {
    err := g.repo.Pull(&git.PullOptions{
        Auth: g.auth,
    })
    return err
}
```

## Verificación
- [ ] `P` muestra spinner "Pushing..."
- [ ] Push success: mensaje de éxito + refresh
- [ ] Push error: mensaje de error descriptivo
- [ ] `p` muestra spinner "Pulling..."
- [ ] Pull success: mensaje + refresh status/log
- [ ] Pull error: mensaje de error
- [ ] UI no crashea con repos públicos sin auth
