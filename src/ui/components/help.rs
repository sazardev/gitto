use ratatui::{
    layout::Rect,
    style::Style,
    text::{Line, Span},
    widgets::{Block, BorderType, Borders, Clear, Paragraph},
    Frame,
};
use crate::core::state::AppState;
use crate::ui::keybind;

pub fn render_help(state: &AppState, f: &mut Frame, area: Rect) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let popup_width = 56;
    let popup_height = 20;
    let popup_x = (area.width.saturating_sub(popup_width)) / 2;
    let popup_y = (area.height.saturating_sub(popup_height)) / 2;

    let popup_area = Rect {
        x: popup_x,
        y: popup_y,
        width: popup_width,
        height: popup_height,
    };

    f.render_widget(Clear, popup_area);

    let block = Block::default()
        .borders(Borders::ALL)
        .border_type(BorderType::Rounded)
        .border_style(theme.border_active())
        .title(" Ayuda [?] ")
        .style(Style::default().bg(theme.surface));

    let inner = block.inner(popup_area);
    f.render_widget(block, popup_area);

    let kb = &state.config.keybindings;
    let accent = theme.accent_primary;
    let muted = theme.muted;

    let lines: Vec<Line> = vec![
        Line::from(Span::styled("  Navegacion", Style::default().fg(accent).add_modifier(ratatui::style::Modifier::BOLD))),
        Line::from(""),
        entry(&keybind::binding_label(&kb.up), "Arriba", accent, muted),
        entry(&keybind::binding_label(&kb.down), "Abajo", accent, muted),
        entry("<Tab>", "Cambiar vista (Status/Log/Diff)", accent, muted),
        entry(&keybind::binding_label(&kb.diff_view), "Abrir diff", accent, muted),
        entry(&keybind::binding_label(&kb.back), "Cerrar / volver atras", accent, muted),
        Line::from(""),
        Line::from(Span::styled("  Acciones", Style::default().fg(accent).add_modifier(ratatui::style::Modifier::BOLD))),
        Line::from(""),
        entry(&keybind::binding_label(&kb.toggle_stage), "Stage / Unstage archivo", accent, muted),
        entry(&keybind::binding_label(&kb.stage), "Stage archivo", accent, muted),
        entry(&keybind::binding_label(&kb.unstage), "Unstage archivo", accent, muted),
        entry(&keybind::binding_label(&kb.commit), "Crear commit", accent, muted),
        entry(&keybind::binding_label(&kb.log_view), "Ver historial", accent, muted),
        entry(&keybind::binding_label(&kb.push), "Push al remoto", accent, muted),
        entry(&keybind::binding_label(&kb.pull), "Pull del remoto", accent, muted),
        Line::from(""),
        Line::from(Span::styled("  General", Style::default().fg(accent).add_modifier(ratatui::style::Modifier::BOLD))),
        Line::from(""),
        entry(&keybind::binding_label(&kb.command_palette), "Paleta de comandos", accent, muted),
        entry(&keybind::binding_label(&kb.search), "Buscar archivo (filtro)", accent, muted),
        entry(&keybind::binding_label(&kb.help), "Esta ayuda", accent, muted),
        entry("Ctrl+C", "Salir", accent, muted),
    ];

    let paragraph = Paragraph::new(lines);
    f.render_widget(paragraph, inner);
}

fn entry(key: &str, desc: &str, accent: Color, muted: Color) -> Line<'static> {
    let key_str = key.to_string();
    let desc_str = desc.to_string();
    Line::from(vec![
        Span::styled(format!("    {:<16}", key_str), Style::default().fg(accent)),
        Span::styled(desc_str, Style::default().fg(muted)),
    ])
}

use ratatui::style::Color;
