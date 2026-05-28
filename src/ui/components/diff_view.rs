use ratatui::{
    layout::Rect,
    style::{Modifier, Style},
    text::{Line, Span},
    widgets::{Block, BorderType, Borders, Paragraph},
    Frame,
};
use crate::core::state::AppState;

pub fn render_diff(state: &AppState, f: &mut Frame, area: Rect) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let title = state
        .selected_file()
        .map(|f| format!(" Diff: {} ", f.path))
        .unwrap_or_else(|| String::from(" Diff "));

    let block = Block::default()
        .borders(Borders::ALL)
        .border_type(BorderType::Rounded)
        .border_style(theme.border_active())
        .title(title);

    let inner_area = block.inner(area);
    f.render_widget(block, area);

    let content = state.diff_content.as_deref().unwrap_or("");

    if content.is_empty() {
        f.render_widget(
            Paragraph::new("Sin cambios en este archivo").style(theme.text_dim()),
            inner_area,
        );
        return;
    }

    let visible_rows = inner_area.height as usize;
    let total_lines = content.lines().count();
    let max_offset = total_lines.saturating_sub(visible_rows);
    let offset = state.scroll_offset.min(max_offset);

    let mut lines: Vec<Line> = Vec::new();

    for (i, line_str) in content.lines().enumerate() {
        if i < offset || i >= offset.saturating_add(visible_rows) {
            continue;
        }

        let line_num = i + 1;
        let num_str = format!("{:>4} ", line_num);
        let num_span = Span::styled(num_str, Style::default().fg(theme.muted));

        let highlighted = i == offset.saturating_add(visible_rows / 2);

        let (color, add_mod) = if line_str.starts_with('+') {
            (theme.success, Modifier::BOLD)
        } else if line_str.starts_with('-') {
            (theme.danger, Modifier::BOLD)
        } else if line_str.starts_with('@') && line_str.contains("@@") {
            (theme.accent_primary, Modifier::empty())
        } else {
            (theme.muted, Modifier::empty())
        };

        let mut content_style = Style::default().fg(color).add_modifier(add_mod);
        if highlighted {
            content_style = content_style.bg(theme.accent_primary).fg(ratatui::style::Color::Black);
        }

        let content_span = Span::styled(line_str, content_style);
        lines.push(Line::from(vec![num_span, content_span]));
    }

    f.render_widget(Paragraph::new(lines), inner_area);
}
