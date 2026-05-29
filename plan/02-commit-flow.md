# Plan: Commit Flow Completo

## Problema Actual
CommitView existe pero:
- No ejecuta el commit real via gogit
- No muestra feedback de éxito/error
- No hace refresh del log ni del status después de commit

## Flujo Esperado
1. Usuario presiona `c` → aparece CommitView (input flotante)
2. Usuario escribe mensaje de commit
3. Usuario presiona `<Enter>` para confirmar
4. Se muestra spinner mientras se ejecuta
5. On success: input desaparece, status/log se refrescan, mensaje de éxito
6. On error: mensaje de error visible, input permanece

## Implementación

### 1. Message types
**File:** `internal/ui/messages.go`

```go
type CommitStarted struct{}

type CommitSuccess struct {
    hash string
}

type CommitError struct {
    error
}

type CloseCommitView struct{}
```

### 2. Update CommitView - add Enter handling
**File:** `internal/ui/views/commit.go`

```go
func (v CommitView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, key.NewBinding(tea.KeyEnter)):
            return v, v.handleSubmit()
        case key.Matches(msg, key.NewBinding(tea.KeyEsc)):
            return v, func() tea.Msg { return CloseCommitView{} }
        case key.Matches(msg, key.NewBinding(tea.KeyBackspace)):
            v.textInput.SetCursorToEnd()
            v.textInput.DeleteChar()
        case isRuneKey(msg):
            v.textInput.Insert(string(msg.Runes[0]))
        }
    }
    return v, nil
}

func (v CommitView) handleSubmit() tea.Cmd {
    if v.textInput.Value() == "" {
        return nil
    }
    return func() tea.Msg {
        hash, err := v.git.Commit(v.textInput.Value())
        if err != nil {
            return CommitError{err}
        }
        return CommitSuccess{hash: hash}
    }
}
```

### 3. MainModel.Update - handle CommitView mode
**File:** `internal/ui/model.go`

```go
case CommitSuccess:
    m.viewState = StatusView
    m.commitMessage = ""
    m.lastMessage = fmt.Sprintf("Committed: %s", msg.hash[:7])
    return m, func() tea.Msg { return RefreshStatus{} }

case CommitError:
    m.lastError = msg.error.Error()
    return m, nil

case CloseCommitView:
    m.viewState = StatusView
    m.commitMessage = ""
    return m, nil
```

### 4. Show spinner during CommitStarted
**File:** `internal/ui/model.go`

```go
case CommitStarted:
    m.isLoading = true
    m.loadingMessage = "Committing..."
    return m, nil
```

### 5. Add RefreshStatus message + handler
**File:** `internal/ui/messages.go`

```go
type RefreshStatus struct{}
```

In Update:
```go
case RefreshStatus:
    return m, func() tea.Msg {
        status, err := m.git.Status()
        if err != nil {
            return StatusError{err}
        }
        return StatusUpdated{status}
    }
```

### 6. Update MainModel.View - show loading spinner
**File:** `internal/ui/model.go`

```go
func (m MainModel) View() string {
    var s strings.Builder
    
    if m.isLoading {
        s.WriteString(m.spinner.View())
        s.WriteString(" ")
        s.WriteString(m.loadingMessage)
        s.WriteString("\n\n")
    }
    
    // render current view based on m.viewState
    switch m.viewState {
    case StatusView:
        s.WriteString(m.statusView.View(m.status, m.cursor))
    case CommitView:
        s.WriteString(m.statusView.View(m.status, m.cursor))
        s.WriteString("\n")
        s.WriteString(m.commitView.View())
    }
    
    return s.String()
}
```

## Verificación
- [ ] `c` abre input de commit
- [ ] Escribir mensaje funciona
- [ ] `<Enter>` ejecuta commit con spinner
- [ ] Éxito: input cierra, status/log refrescan
- [ ] Error: mensaje de error visible
- [ ] `<Esc>` cancela sin commitear
