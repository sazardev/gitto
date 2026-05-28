use ratatui::{
    layout::Rect,
    style::Style,
    text::{Line, Span},
    widgets::{Block, BorderType, Borders, Paragraph},
    Frame,
};
use crate::core::state::AppState;

pub fn render_log(state: &AppState, f: &mut Frame, area: Rect) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let block = Block::default()
        .borders(Borders::ALL)
        .border_type(BorderType::Rounded)
        .border_style(theme.border_active())
        .title(" Historial ");

    let inner_area = block.inner(area);
    f.render_widget(block, area);

    if state.commits.is_empty() {
        let msg = Paragraph::new("Sin commits")
            .style(theme.text_dim());
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
        let first_line = commit.message.lines().next().unwrap_or("");

        let time_str = relative_time(commit.timestamp);
        let author = &commit.author;

        let mut spans = Vec::new();

        if is_selected {
            spans.push(Span::styled(
                format!(" {} ", short_hash),
                Style::default().fg(theme.success).bg(theme.accent_primary),
            ));
            spans.push(Span::styled(
                format!(" {}", first_line),
                Style::default().fg(Color::Black).bg(theme.accent_primary),
            ));
            spans.push(Span::styled(
                format!("  {}  {}", author, time_str),
                Style::default().fg(theme.muted).bg(theme.accent_primary),
            ));
        } else {
            spans.push(Span::styled(
                format!(" {} ", short_hash),
                Style::default().fg(theme.success),
            ));
            spans.push(Span::styled(
                format!(" {}", first_line),
                theme.text(),
            ));
            spans.push(Span::styled(
                format!("  {}  {}", author, time_str),
                theme.text_dim(),
            ));
        }

        let line_width = inner_area.width as usize;
        let padded = pad_to_width(&spans, line_width, theme.bg_base);
        lines.push(Line::from(padded));
    }

    let paragraph = Paragraph::new(lines);
    f.render_widget(paragraph, inner_area);
}

use ratatui::style::Color;

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

fn pad_to_width(spans: &[Span<'_>], width: usize, bg: Color) -> Vec<Span<'static>> {
    let current_len: usize = spans.iter().map(|s| s.content.chars().count()).sum();
    let mut result: Vec<Span<'static>> = spans.iter().map(|s| Span::styled(s.content.to_string(), s.style)).collect();
    if current_len < width {
        let padding = width - current_len;
        result.push(Span::styled(
            " ".repeat(padding),
            Style::default().bg(bg),
        ));
    }
    result
}
