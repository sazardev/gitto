use ratatui::{
    layout::Rect,
    style::Style,
    text::{Line, Span},
    widgets::{Block, BorderType, Borders, Clear, List, ListItem, Paragraph},
    Frame,
};
use crate::core::state::AppState;

struct Cmd {
    name: &'static str,
    description: &'static str,
    shortcut: &'static str,
}

const COMMANDS: &[Cmd] = &[
    Cmd { name: "status", description: "Ir a la vista de estado", shortcut: "Esc" },
    Cmd { name: "log", description: "Mostrar historial de commits", shortcut: "l" },
    Cmd { name: "commit", description: "Crear un nuevo commit", shortcut: "c" },
    Cmd { name: "push", description: "Enviar cambios al remoto", shortcut: "P" },
    Cmd { name: "pull", description: "Traer cambios del remoto", shortcut: "p" },
    Cmd { name: "quit", description: "Salir de gitto", shortcut: "q" },
];

pub fn render_command_palette(state: &AppState, f: &mut Frame, area: Rect) {
    let theme = crate::ui::theme::Theme::from_hex_strings(&state.config.theme);

    let matches = fuzzy_filter(&state.command_input, COMMANDS);
    let popup_height = (matches.len() + 3).min(14) as u16;
    let popup_width = 62;
    let popup_x = (area.width.saturating_sub(popup_width)) / 2;
    let popup_y = (area.height.saturating_sub(popup_height)) / 4;

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
        .style(Style::default().bg(theme.surface));

    let inner = block.inner(popup_area);
    f.render_widget(block, popup_area);

    let chunks = ratatui::layout::Layout::default()
        .direction(ratatui::layout::Direction::Vertical)
        .constraints([
            ratatui::layout::Constraint::Length(1),
            ratatui::layout::Constraint::Min(0),
        ])
        .split(inner);

    let prompt = format!("> {}", state.command_input);
    let input_paragraph = Paragraph::new(Span::styled(
        prompt,
        Style::default().fg(theme.accent_primary),
    ));
    f.render_widget(input_paragraph, chunks[0]);

    let items: Vec<ListItem> = matches
        .iter()
        .map(|cmd| {
            let name_span = Span::styled(
                cmd.name,
                Style::default().fg(theme.accent_primary),
            );
            let desc_span = Span::styled(
                format!("  {}", cmd.description),
                theme.text_dim(),
            );
            let shortcut_span = Span::styled(
                format!("  [{}]", cmd.shortcut),
                Style::default().fg(theme.muted),
            );
            ListItem::new(Line::from(vec![name_span, desc_span, shortcut_span]))
        })
        .collect();

    let list = List::new(items);
    f.render_widget(list, chunks[1]);

    f.set_cursor_position((popup_x + 3 + state.command_input.len() as u16, popup_y + 2));
}

fn fuzzy_filter<'a>(query: &str, cmds: &'a [Cmd]) -> Vec<&'a Cmd> {
    if query.is_empty() {
        return cmds.iter().collect();
    }
    let q = query.to_lowercase();
    cmds.iter()
        .filter(|cmd| cmd.name.contains(&q) || cmd.description.to_lowercase().contains(&q))
        .collect()
}
