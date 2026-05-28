use ratatui::{
    layout::Rect,
    style::Style,
    text::{Line, Span},
    widgets::{Block, Borders, Clear, Paragraph},
    Frame,
};
use crate::core::state::AppState;

pub fn render_commit_input(state: &AppState, f: &mut Frame, area: Rect) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let popup_height = 5;
    let popup_area = Rect {
        x: 2,
        y: area.height.saturating_sub(popup_height + 1),
        width: area.width.saturating_sub(4),
        height: popup_height,
    };

    f.render_widget(Clear, popup_area);

    let block = Block::default()
        .borders(Borders::ALL)
        .border_style(Style::default().fg(theme.accent_primary))
        .title(format!(
            " Mensaje de commit [{}] ",
            state.commit_message.chars().count()
        ))
        .style(Style::default().bg(theme.surface));

    let inner = block.inner(popup_area);
    f.render_widget(block, popup_area);

    let lines: Vec<Line> = state
        .commit_message
        .lines()
        .enumerate()
        .map(|(i, line)| {
            if i == 0 {
                Line::from(Span::styled(
                    format!("> {}", line),
                    Style::default().fg(theme.accent_primary),
                ))
            } else {
                Line::from(Span::styled(
                    format!("  {}", line),
                    theme.text_dim(),
                ))
            }
        })
        .collect();

    if lines.is_empty() {
        let hint = Paragraph::new(Span::styled(
            "> ",
            Style::default().fg(theme.muted),
        ));
        f.render_widget(hint, inner);
    } else {
        let paragraph = Paragraph::new(lines);
        f.render_widget(paragraph, inner);
    }

    let cursor_x = 4 + state.commit_message.chars().count().min(inner.width as usize - 4);
    f.set_cursor_position((
        popup_area.x + 2 + cursor_x as u16,
        popup_area.y + 1,
    ));
}
