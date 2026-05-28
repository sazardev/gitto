use ratatui::{
    layout::Rect,
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, BorderType, Borders, Clear, Paragraph},
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

    if state.is_loading {
        render_skeleton(state, f, inner_area, &theme);
        return;
    }

    let all_files: Vec<&crate::core::models::FileEntry> = if state.show_search && !state.search_query.is_empty() {
        let q = state.search_query.to_lowercase();
        state.all_files().into_iter().filter(|f| {
            f.path.to_lowercase().contains(&q)
                || f.old_path.as_ref().map_or(false, |o| o.to_lowercase().contains(&q))
        }).collect()
    } else {
        state.all_files()
    };

    if all_files.is_empty() {
        let msg = if state.show_search && !state.search_query.is_empty() {
            format!("  Sin resultados para \"{}\"", state.search_query)
        } else {
            String::from("  Sin cambios en el repositorio")
        };
        f.render_widget(Paragraph::new(msg).style(theme.text_dim()), inner_area);
        return;
    }

    let visual_map = state.visual_map_for(&all_files);

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
    render_file_list(state, &all_files, &visual_map, f, chunks[1], content_height);
}

fn render_skeleton(state: &AppState, f: &mut Frame, area: Rect, theme: &crate::ui::theme::Theme) {
    let msg = format!("  {} ...", state.loading_message);
    f.render_widget(Paragraph::new(msg).style(theme.text_dim()), area);
}

pub fn render_search_bar(state: &AppState, f: &mut Frame, area: Rect) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let popup_area = Rect {
        x: area.width.saturating_sub(42) / 2,
        y: 2,
        width: 42,
        height: 3,
    };

    f.render_widget(Clear, popup_area);

    let block = Block::default()
        .borders(Borders::ALL)
        .border_type(BorderType::Rounded)
        .border_style(theme.border_active())
        .style(Style::default().bg(theme.surface));

    f.render_widget(block, popup_area);

    let prompt = format!("/{}", state.search_query);
    let inner = Rect {
        x: popup_area.x + 2,
        y: popup_area.y + 1,
        width: popup_area.width.saturating_sub(4),
        height: 1,
    };
    f.render_widget(
        Paragraph::new(Span::styled(prompt, Style::default().fg(theme.accent_primary))),
        inner,
    );

    f.set_cursor_position((inner.x + state.search_query.len() as u16 + 1, inner.y));
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
    visual_map: &[Option<usize>],
    f: &mut Frame,
    area: Rect,
    content_height: usize,
) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let total_visual = visual_map.len();
    if total_visual == 0 {
        return;
    }

    let start = state.scroll_offset.min(total_visual.saturating_sub(1));
    let end = (start + content_height).min(total_visual);

    let flash_set: std::collections::HashSet<&str> = state
        .flash_files
        .iter()
        .map(|(p, _)| p.as_str())
        .collect();

    let mut lines: Vec<Line> = Vec::new();
    let mut last_section: Option<&str> = None;

    for v in start..end {
        match visual_map[v] {
            Some(file_idx) => {
                let file = all_files[file_idx];
                let section = section_name(&file.state);
                if last_section != Some(section) {
                    if !lines.is_empty() {
                        lines.push(Line::from(""));
                    }
                    let header = section_header(section, &theme);
                    lines.push(header);
                    last_section = Some(section);
                }

                let is_selected = file_idx == state.selected_index;
                let is_flashing = flash_set.contains(file.path.as_str());
                lines.push(file_line(file, is_selected, is_flashing, &theme));
            }
            None => {}
        }
    }

    let paragraph = Paragraph::new(lines);
    f.render_widget(paragraph, area);
}


const fn section_name(state: &FileState) -> &'static str {
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
        Style::default().fg(color).add_modifier(Modifier::BOLD),
    ))
}

fn file_line(
    file: &crate::core::models::FileEntry,
    selected: bool,
    flashing: bool,
    theme: &crate::ui::theme::Theme,
) -> Line<'static> {
    let icon = status_icon(&file.state);
    let display_path = match &file.old_path {
        Some(old) => format!("{} -> {}", old, file.path),
        None => file.path.clone(),
    };
    let file_icon = extension_icon(&file.path);
    let text = format!("    {} {} {}", icon, file_icon, display_path);

    let style = if selected {
        theme.selected_style()
    } else if flashing {
        Style::default().bg(theme.success).add_modifier(Modifier::BOLD)
    } else {
        file_state_style(&file.state, theme)
    };

    Line::from(Span::styled(text, style))
}

fn file_state_style(state: &FileState, theme: &crate::ui::theme::Theme) -> Style {
    match state {
        FileState::Staged => Style::default().fg(theme.success),
        FileState::Unstaged => Style::default().fg(theme.muted),
        FileState::Untracked => Style::default().fg(theme.accent_primary),
        FileState::Renamed => Style::default().fg(theme.danger),
        FileState::Conflicted => Style::default().fg(theme.danger).add_modifier(Modifier::BOLD),
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
