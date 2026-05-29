use ratatui::{
    layout::Rect,
    style::{Color, Style},
    text::{Line, Span},
    widgets::Paragraph,
    Frame,
};
use crate::core::state::AppState;
use crate::ui::keybind;

pub fn render_footer(state: &AppState, f: &mut Frame, area: Rect) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);
    let kb = &state.config.keybindings;

    let up_label = keybind::binding_label(&kb.up);
    let down_label = keybind::binding_label(&kb.down);
    let toggle_label = keybind::binding_label(&kb.toggle_stage);
    let unstage_label = keybind::binding_label(&kb.unstage);
    let commit_label = keybind::binding_label(&kb.commit);
    let log_label = keybind::binding_label(&kb.log_view);
    let search_label = keybind::binding_label(&kb.search);
    let diff_label = keybind::binding_label(&kb.diff_view);
    let back_label = keybind::binding_label(&kb.back);
    let quit_label = keybind::binding_label(&kb.quit);
    let split_h_label = keybind::binding_label(&kb.toggle_split_horizontal);
    let split_v_label = keybind::binding_label(&kb.toggle_split_vertical);
    let zen_label = keybind::binding_label(&kb.toggle_zen);

    let mut spans: Vec<Span> = Vec::new();
    spans.push(Span::raw(" "));

    match state.view {
        crate::core::state::View::Status => {
            pair_key(&mut spans, &up_label, &down_label, "Nav", theme.accent_primary, theme.muted);
            single_key(&mut spans, &toggle_label, "Stage", theme.success, theme.muted);
            single_key(&mut spans, &unstage_label, "Unstage", theme.danger, theme.muted);
            single_key(&mut spans, &commit_label, "Commit", theme.accent_primary, theme.muted);
            single_key(&mut spans, &log_label, "Log", theme.accent_primary, theme.muted);
            single_key(&mut spans, &search_label, "Buscar", theme.accent_primary, theme.muted);
        }
        crate::core::state::View::Log => {
            pair_key(&mut spans, &up_label, &down_label, "Nav", theme.accent_primary, theme.muted);
            single_key(&mut spans, &diff_label, "Ver diff", theme.accent_primary, theme.muted);
            single_key(&mut spans, &back_label, "Volver", theme.danger, theme.muted);
        }
        crate::core::state::View::Diff => {
            pair_key(&mut spans, &up_label, &down_label, "Scroll", theme.accent_primary, theme.muted);
            single_key(&mut spans, &back_label, "Volver", theme.danger, theme.muted);
        }
    }

    spans.push(Span::raw(" "));
    spans.push(Span::styled("│", theme.muted));

    match state.layout {
        crate::core::state::Layout::Zen => {
            single_key(&mut spans, &split_h_label, "Split H", theme.accent_primary, theme.muted);
            single_key(&mut spans, &split_v_label, "Split V", theme.accent_primary, theme.muted);
        }
        crate::core::state::Layout::SplitHorizontal => {
            single_key(&mut spans, &zen_label, "Zen", theme.success, theme.muted);
            single_key(&mut spans, &split_v_label, "Split V", theme.accent_primary, theme.muted);
        }
        crate::core::state::Layout::SplitVertical => {
            single_key(&mut spans, &zen_label, "Zen", theme.success, theme.muted);
            single_key(&mut spans, &split_h_label, "Split H", theme.accent_primary, theme.muted);
        }
    }

    spans.push(Span::raw(" "));
    spans.push(Span::styled("│", theme.muted));
    single_key(&mut spans, &quit_label, "Salir", theme.danger, theme.muted);

    let paragraph = Paragraph::new(Line::from(spans))
        .style(Style::default().bg(theme.surface));
    f.render_widget(paragraph, area);
}

fn pair_key<'a>(spans: &mut Vec<Span<'a>>, k1: &'a str, k2: &'a str, label: &'a str, key_color: Color, label_color: Color) {
    spans.push(Span::raw(" "));
    spans.push(Span::styled(k1, Style::default().fg(key_color)));
    spans.push(Span::raw("/"));
    spans.push(Span::styled(k2, Style::default().fg(key_color)));
    spans.push(Span::raw(" "));
    spans.push(Span::styled(label, Style::default().fg(label_color)));
}

fn single_key<'a>(spans: &mut Vec<Span<'a>>, key: &'a str, label: &'a str, key_color: Color, label_color: Color) {
    spans.push(Span::raw(" "));
    spans.push(Span::styled(key, Style::default().fg(key_color)));
    spans.push(Span::raw(" "));
    spans.push(Span::styled(label, Style::default().fg(label_color)));
}
