use ratatui::{
    layout::Rect,
    style::{Modifier, Style},
    text::{Line, Span},
    widgets::{Block, BorderType, Borders, Paragraph},
    Frame,
};
use crate::core::state::AppState;
use crate::core::models::FileState;

pub fn render_changes_overview(state: &AppState, f: &mut Frame, area: Rect) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let block = Block::default()
        .borders(Borders::ALL)
        .border_type(BorderType::Rounded)
        .border_style(theme.border_active())
        .title(" Cambios ");

    let inner_area = block.inner(area);
    f.render_widget(block, area);

    let staged_count = state.staged_count();
    let unstaged_count = state.unstaged_or_untracked_count();

    let mut lines: Vec<Line> = Vec::new();

    lines.push(Line::from(vec![
        Span::styled("  ───────────────────────────────────────────────", theme.text_dim())
    ]));

    lines.push(Line::from(vec![
        Span::raw("  "),
        Span::styled("Staged", Style::default().fg(theme.success).add_modifier(Modifier::BOLD)),
        Span::raw("                    "),
        Span::styled(format!("{} files", staged_count), theme.text()),
    ]));

    lines.push(render_staged_files(state, &theme));

    lines.push(Line::from(vec![
        Span::styled("  ───────────────────────────────────────────────", theme.text_dim())
    ]));

    lines.push(Line::from(vec![
        Span::raw("  "),
        Span::styled("Unstaged", Style::default().fg(theme.muted).add_modifier(Modifier::BOLD)),
        Span::raw("                  "),
        Span::styled(format!("{} files", unstaged_count), theme.text()),
    ]));

    lines.push(render_unstaged_files(state, &theme));

    lines.push(Line::from(vec![
        Span::styled("  ───────────────────────────────────────────────", theme.text_dim())
    ]));

    let total_changes = staged_count + unstaged_count;
    lines.push(Line::from(vec![
        Span::raw("  "),
        Span::styled("Total", theme.accent()),
        Span::raw("                            "),
        Span::styled(format!("{} changes", total_changes), theme.accent()),
    ]));

    let paragraph = Paragraph::new(lines);
    f.render_widget(paragraph, inner_area);
}

fn render_staged_files(state: &AppState, theme: &crate::ui::theme::Theme) -> Line<'static> {
    if state.staged.is_empty() {
        return Line::from(vec![Span::styled("    No hay archivos staged", theme.text_dim())]);
    }

    let file_names: Vec<String> = state.staged.iter()
        .take(5)
        .map(|f| {
            let icon = match f.state {
                FileState::Staged => "+",
                _ => "~",
            };
            format!("{}{}", icon, f.path.split('/').last().unwrap_or(&f.path))
        })
        .collect();

    let more = if state.staged.len() > 5 {
        format!(" y {} más...", state.staged.len() - 5)
    } else {
        String::new()
    };

    Line::from(vec![Span::styled(
        format!("    {}{}", file_names.join("  "), more),
        theme.success,
    )])
}

fn render_unstaged_files(state: &AppState, theme: &crate::ui::theme::Theme) -> Line<'static> {
    if state.files.is_empty() {
        return Line::from(vec![Span::styled("    No hay cambios unstaged", theme.text_dim())]);
    }

    let file_names: Vec<String> = state.files.iter()
        .take(5)
        .map(|f| {
            let icon = match f.state {
                FileState::Unstaged => "~",
                FileState::Untracked => "?",
                _ => "~",
            };
            format!("{}{}", icon, f.path.split('/').last().unwrap_or(&f.path))
        })
        .collect();

    let more = if state.files.len() > 5 {
        format!(" y {} más...", state.files.len() - 5)
    } else {
        String::new()
    };

    Line::from(vec![Span::styled(
        format!("    {}{}", file_names.join("  "), more),
        theme.muted,
    )])
}

pub fn render_flow(state: &AppState, f: &mut Frame, area: Rect) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let block = Block::default()
        .borders(Borders::ALL)
        .border_type(BorderType::Rounded)
        .border_style(theme.border_active())
        .title(" Flujo ");

    let inner_area = block.inner(area);
    f.render_widget(block, area);

    if state.commits.is_empty() {
        let msg = Paragraph::new("Sin commits").style(theme.text_dim());
        f.render_widget(msg, inner_area);
        return;
    }

    let visible_rows = inner_area.height as usize;
    let max_offset = state.commits.len().saturating_sub(visible_rows);
    let offset = state.scroll_offset.min(max_offset);

    let mut lines: Vec<Line> = Vec::new();

    for (i, commit) in state.commits.iter().enumerate().skip(offset).take(visible_rows) {
        let is_selected = i == state.selected_index;
        let short_hash = &commit.hash[..commit.hash.len().min(7)];

        let icon = if i == 0 { "●" } else if i == state.commits.len().saturating_sub(1) { "○" } else { "│" };

        let time_str = relative_time(commit.timestamp);

        if is_selected {
            lines.push(Line::from(vec![
                Span::styled(format!("  {} ", icon), Style::default().fg(theme.accent_primary)),
                Span::styled(format!("{} ", short_hash), theme.success),
                Span::styled(format!("{} ", commit.message.lines().next().unwrap_or("")), theme.text()),
                Span::styled(time_str, theme.text_dim()),
            ]));
        } else {
            lines.push(Line::from(vec![
                Span::styled(format!("  {} ", icon), theme.muted),
                Span::styled(format!("{} ", short_hash), Style::default().fg(theme.success)),
                Span::styled(format!("{} ", commit.message.lines().next().unwrap_or("")), theme.text_dim()),
                Span::styled(time_str, theme.text_dim()),
            ]));
        }
    }

    let paragraph = Paragraph::new(lines);
    f.render_widget(paragraph, inner_area);
}

fn relative_time(timestamp: i64) -> String {
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs() as i64;
    let diff = now.saturating_sub(timestamp);

    if diff < 60 {
        format!("hace {}s", diff)
    } else if diff < 3600 {
        format!("hace {}min", diff / 60)
    } else if diff < 86400 {
        format!("hace {}h", diff / 3600)
    } else if diff < 604800 {
        format!("hace {}d", diff / 86400)
    } else {
        format!("hace {}semanas", diff / 604800)
    }
}