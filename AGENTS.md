# gitto — Zen Git TUI

## Architecture
- **Pattern:** Ports & Adapters (Hexagonal)
  - `core/` — domain models, business rules (no UI/Git deps)
  - `ports/` — traits: `GitProvider`, `ConfigProvider`
  - `adapters/` — implementations (git2, fs config)
  - `ui/` — ratatui loop, widgets, event handling
- **Stack:** ratatui, crossterm, git2 (libgit2 bindings), tokio, serde + toml, anyhow, thiserror
- **Rust edition:** 2024

## Critical Constraints
- Never use `std::process::Command` for Git read operations — use `git2` exclusively
- MVP scope: Status, Stage/Unstage, Commit, Push/Pull, Log, Diff (see SPEC.md §6-7)
- Out of scope for v0.1.0: interactive rebase, merge conflict resolution, multi-repo

## Visual Design
- Read `.agents/skills/gitto-tui-design/SKILL.md` before any UI work
- Nerd Fonts required for icons
- Rounded panel borders (╭╰╮╯), pastel color palette, double-buffered rendering
- Navigation: `hjkl` movement, `:` for command palette, `Esc` as universal back

## Conventions
- Docs and comments are written in Spanish
- Rust best practices: review `.agents/skills/rust-best-practices/SKILL.md`
- Use `anyhow` for binary-level errors, `thiserror` for library-level
