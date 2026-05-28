use ratatui::{
    layout::Rect,
    style::{Color, Style},
    text::{Line, Span},
    widgets::Paragraph,
    Frame,
};
use crate::core::state::AppState;

pub fn render_footer(state: &AppState, f: &mut Frame, area: Rect) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let mut spans: Vec<Span> = Vec::new();
    spans.push(Span::raw(" "));

    match state.view {
        crate::core::state::View::Status => {
            add_key(&mut spans, "j/k", "Navegar", theme.accent_primary, theme.muted);
            add_key(&mut spans, "s", "Stage", theme.success, theme.muted);
            add_key(&mut spans, "u", "Unstage", theme.danger, theme.muted);
            add_key(&mut spans, "c", "Commit", theme.accent_primary, theme.muted);
            add_key(&mut spans, "l", "Log", theme.accent_primary, theme.muted);
            add_key(&mut spans, "P/p", "Push/Pull", theme.accent_primary, theme.muted);
            add_key(&mut spans, ":", "Comandos", theme.muted, theme.muted);
        }
        crate::core::state::View::Log => {
            add_key(&mut spans, "j/k", "Navegar", theme.accent_primary, theme.muted);
            add_key(&mut spans, "Enter", "Ver diff", theme.accent_primary, theme.muted);
            add_key(&mut spans, "Esc/q", "Volver", theme.danger, theme.muted);
        }
        crate::core::state::View::Diff => {
            add_key(&mut spans, "j/k", "Scroll", theme.accent_primary, theme.muted);
            add_key(&mut spans, "Esc/q", "Volver", theme.danger, theme.muted);
        }
    }

    spans.push(Span::raw(" "));
    add_key(&mut spans, "q", "Salir", theme.danger, theme.muted);

    let paragraph = Paragraph::new(Line::from(spans))
        .style(Style::default().bg(theme.surface));
    f.render_widget(paragraph, area);
}

fn add_key<'a>(spans: &mut Vec<Span<'a>>, key: &'a str, label: &'a str, key_color: Color, label_color: Color) {
    spans.push(Span::raw(" "));
    spans.push(Span::styled(key, Style::default().fg(key_color)));
    spans.push(Span::raw(" "));
    spans.push(Span::styled(label, Style::default().fg(label_color)));
}
