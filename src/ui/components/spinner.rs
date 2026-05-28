use ratatui::{
    layout::Rect,
    style::Style,
    text::Span,
    widgets::Paragraph,
    Frame,
};
use crate::core::state::AppState;

const SPINNER_FRAMES: [&str; 10] = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];

pub fn render_spinner(state: &AppState, f: &mut Frame, area: Rect, tick: usize) {
    if !state.is_loading {
        return;
    }

    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);
    let frame = SPINNER_FRAMES[tick % SPINNER_FRAMES.len()];
    let text = format!(" {}  {}", frame, state.loading_message);

    let spinner_area = Rect {
        x: 1,
        y: area.height.saturating_sub(1),
        width: text.len() as u16,
        height: 1,
    };

    let paragraph = Paragraph::new(Span::styled(
        text,
        Style::default().fg(theme.bg_base).bg(theme.accent_primary),
    ));

    f.render_widget(paragraph, spinner_area);
}
