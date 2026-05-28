use ratatui::{
    layout::Rect,
    style::{Color, Style},
    text::{Line, Span},
    widgets::{Block, BorderType, Borders, Paragraph},
    Frame,
};
use crate::core::models::FileState;
use crate::core::state::AppState;

pub fn render_status(state: &AppState, f: &mut Frame, area: Rect) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let inner = Block::default()
        .borders(Borders::ALL)
        .border_type(BorderType::Rounded)
        .border_style(if state.show_command_palette || state.show_commit_input {
            theme.border_inactive()
        } else {
            theme.border_active()
        });

    let inner_area = inner.inner(area);
    f.render_widget(inner, area);

    let all_files = state.all_files();

    if all_files.is_empty() && !state.is_loading {
        let msg = Paragraph::new("  Sin cambios en el repositorio")
            .style(theme.text_dim());
        f.render_widget(msg, inner_area);
        return;
    }

    let header_height = if state.branch.is_some() { 3 } else { 1 };
    let content_height = inner_area.height.saturating_sub(header_height).max(1) as usize;

    let chunks = ratatui::layout::Layout::default()
        .direction(ratatui::layout::Direction::Vertical)
        .constraints([
            ratatui::layout::Constraint::Length(header_height),
            ratatui::layout::Constraint::Min(0),
        ])
        .split(inner_area);

    render_branch_header(state, f, chunks[0], &theme);

    render_file_list(state, &all_files, f, chunks[1], content_height);
}

fn render_branch_header(
    state: &AppState,
    f: &mut Frame,
    area: Rect,
    theme: &crate::ui::theme::Theme,
) {
    if let Some(branch) = &state.branch {
        let mut header_spans = vec![
            Span::raw("  "),
            Span::styled(format!(" {}", branch.name), theme.accent()),
            Span::raw("  "),
        ];

        if branch.ahead > 0 {
            header_spans.push(Span::styled(
                format!(" {} ", branch.ahead),
                Style::default().fg(Color::Black).bg(theme.success),
            ));
            header_spans.push(Span::raw(" "));
        }
        if branch.behind > 0 {
            header_spans.push(Span::styled(
                format!(" {} ", branch.behind),
                Style::default().fg(Color::Black).bg(theme.danger),
            ));
            header_spans.push(Span::raw(" "));
        }

        if let Some(upstream) = &branch.upstream {
            header_spans.push(Span::styled(
                format!(" {} ", upstream),
                theme.text_dim(),
            ));
        }

        let staged_count = state.staged_count();
        let unstaged_count = state.unstaged_or_untracked_count();
        header_spans.push(Span::styled(
            format!("  [+{}] ~{}", staged_count, unstaged_count),
            theme.text_dim(),
        ));

        let paragraph = Paragraph::new(Line::from(header_spans));
        f.render_widget(paragraph, area);
    }
}

fn render_file_list(
    state: &AppState,
    all_files: &[&crate::core::models::FileEntry],
    f: &mut Frame,
    area: Rect,
    content_height: usize,
) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let total = all_files.len();
    if total == 0 {
        return;
    }

    let max_offset = total.saturating_sub(content_height);
    let offset = state.scroll_offset.min(max_offset);
    let visible_end = (offset + content_height).min(total);

    let mut lines: Vec<Line> = Vec::new();
    let mut last_section: Option<&str> = None;

    for i in offset..visible_end {
        let file = all_files[i];
        let section = section_name(&file.state);

        if last_section != Some(section) {
            if !lines.is_empty() {
                lines.push(Line::from(""));
            }
            let header = section_header(section, &theme);
            lines.push(header);
            last_section = Some(section);
        }

        let is_selected = i == state.selected_index;
        lines.push(file_line(file, is_selected, &theme));
    }

    let paragraph = Paragraph::new(lines);
    f.render_widget(paragraph, area);
}

fn section_name(state: &FileState) -> &'static str {
    match state {
        FileState::Staged => "Cambios Staged",
        FileState::Unstaged => "Cambios sin Stagear",
        FileState::Untracked => "Sin Seguimiento",
        FileState::Renamed => "Renombrados",
        FileState::Conflicted => "Conflictos",
    }
}

fn section_header(section: &str, theme: &crate::ui::theme::Theme) -> Line<'static> {
    let color = match section {
        "Cambios Staged" => theme.success,
        "Conflictos" => theme.danger,
        _ => theme.accent_primary,
    };
    Line::from(Span::styled(
        format!("  {}", section),
        Style::default().fg(color),
    ))
}

fn file_line(
    file: &crate::core::models::FileEntry,
    selected: bool,
    theme: &crate::ui::theme::Theme,
) -> Line<'static> {
    let icon = status_icon(&file.state);
    let file_icon = extension_icon(&file.path);
    let text = format!("    {} {} {}", icon, file_icon, file.path);

    if selected {
        Line::from(Span::styled(text, theme.selected_style()))
    } else {
        Line::from(Span::styled(
            text,
            file_state_style(&file.state, theme),
        ))
    }
}

fn file_state_style(state: &FileState, theme: &crate::ui::theme::Theme) -> Style {
    match state {
        FileState::Staged => Style::default().fg(theme.success),
        FileState::Unstaged => Style::default().fg(theme.muted),
        FileState::Untracked => Style::default().fg(theme.accent_primary),
        FileState::Renamed => Style::default().fg(theme.danger),
        FileState::Conflicted => Style::default().fg(theme.danger).add_modifier(ratatui::style::Modifier::BOLD),
    }
}

fn status_icon(state: &FileState) -> &'static str {
    match state {
        FileState::Staged => "+",
        FileState::Unstaged => "~",
        FileState::Untracked => "?",
        FileState::Renamed => ">",
        FileState::Conflicted => "!",
    }
}

fn extension_icon(path: &str) -> &'static str {
    if let Some(ext) = path.rsplit('.').next() {
        match ext {
            "rs" => "\u{e7a8}",
            "toml" => "\u{e6b2}",
            "md" => "\u{e73b}",
            "json" => "\u{e60b}",
            "yaml" | "yml" => "\u{e6b2}",
            "js" => "\u{e781}",
            "ts" => "\u{e7a8}",
            "py" => "\u{e73c}",
            "go" => "\u{e724}",
            "lock" => "\u{e6b2}",
            "gitignore" => "\u{e702}",
            _ => "\u{e716}",
        }
    } else {
        "\u{e716}"
    }
}
