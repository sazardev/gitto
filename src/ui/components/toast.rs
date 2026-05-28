use ratatui::{
    layout::Rect,
    style::{Color, Style},
    text::Span,
    widgets::Paragraph,
    Frame,
};
use crate::core::state::ToastMessage;

const SUCCESS_DURATION: std::time::Duration = std::time::Duration::from_secs(3);
const INFO_DURATION: std::time::Duration = std::time::Duration::from_secs(4);

pub fn should_dismiss(toast: &ToastMessage) -> bool {
    match toast.kind {
        crate::core::state::ToastKind::Error => false,
        crate::core::state::ToastKind::Success => toast.created_at.elapsed() >= SUCCESS_DURATION,
        crate::core::state::ToastKind::Info => toast.created_at.elapsed() >= INFO_DURATION,
    }
}

pub fn render_toast_raw(
    toast: &ToastMessage,
    theme: &crate::ui::theme::Theme,
    f: &mut Frame,
    area: Rect,
) {
    let (icon, color) = match toast.kind {
        crate::core::state::ToastKind::Success => ("  OK  ", theme.success),
        crate::core::state::ToastKind::Error => (" ERROR ", theme.danger),
        crate::core::state::ToastKind::Info => ("  INFO ", theme.accent_primary),
    };

    let text = format!("{} {}", icon, toast.text);
    let width = text.len() as u16;

    let toast_area = Rect {
        x: area.width.saturating_sub(width).saturating_sub(1),
        y: area.height.saturating_sub(1),
        width,
        height: 1,
    };

    let paragraph = Paragraph::new(Span::styled(
        text,
        Style::default().fg(Color::Black).bg(color),
    ));

    f.render_widget(paragraph, toast_area);
}
