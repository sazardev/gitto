use ratatui::style::{Color, Style};

pub struct Theme {
    pub bg_base: Color,
    pub accent_primary: Color,
    pub success: Color,
    pub danger: Color,
    pub muted: Color,
    pub surface: Color,
    pub text: Color,
}

impl Theme {
    pub fn from_hex_strings(hex: &crate::ports::config_provider::ThemeConfig) -> Self {
        let bg = parse_hex(&hex.bg_base).unwrap_or(Color::Rgb(30, 30, 46));
        Self {
            bg_base: bg,
            accent_primary: parse_hex(&hex.accent_primary).unwrap_or(Color::Rgb(203, 166, 247)),
            success: parse_hex(&hex.success).unwrap_or(Color::Rgb(166, 227, 161)),
            danger: parse_hex(&hex.danger).unwrap_or(Color::Rgb(243, 139, 168)),
            muted: parse_hex(&hex.muted).unwrap_or(Color::Rgb(108, 113, 134)),
            surface: Color::Rgb(24, 24, 37),
            text: Color::Rgb(205, 214, 244),
        }
    }

    pub fn bg_style(&self) -> Style {
        Style::default().bg(self.bg_base)
    }

    pub fn text_dim(&self) -> Style {
        Style::default().fg(self.muted)
    }

    pub fn text(&self) -> Style {
        Style::default().fg(self.text)
    }

    pub fn accent(&self) -> Style {
        Style::default().fg(self.accent_primary)
    }

    pub fn heading_style(&self) -> Style {
        Style::default().fg(self.accent_primary)
    }

    pub fn selected_style(&self) -> Style {
        Style::default().bg(self.accent_primary).fg(Color::Black)
    }

    pub fn border_active(&self) -> Style {
        Style::default().fg(self.accent_primary)
    }

    pub fn border_inactive(&self) -> Style {
        Style::default().fg(self.muted)
    }
}

fn parse_hex(hex: &str) -> Option<Color> {
    let hex = hex.trim_start_matches('#');
    if hex.len() != 6 {
        return None;
    }
    let r = u8::from_str_radix(&hex[0..2], 16).ok()?;
    let g = u8::from_str_radix(&hex[2..4], 16).ok()?;
    let b = u8::from_str_radix(&hex[4..6], 16).ok()?;
    Some(Color::Rgb(r, g, b))
}
