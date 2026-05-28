use crossterm::event::{KeyCode, KeyEvent, KeyModifiers};

pub fn matches_binding(key: &KeyEvent, binding: &str) -> bool {
    let binding = binding.trim();
    if binding.is_empty() {
        return false;
    }

    if binding.to_lowercase().starts_with("ctrl+") && binding.len() >= 6 {
        let ch = binding[5..].chars().next().unwrap();
        return key.modifiers.contains(KeyModifiers::CONTROL) && key.code == KeyCode::Char(ch);
    }

    match binding {
        "<Enter>" => key.code == KeyCode::Enter,
        "<Esc>" => key.code == KeyCode::Esc,
        "<Tab>" => key.code == KeyCode::Tab,
        "<Backspace>" => key.code == KeyCode::Backspace,
        "<Up>" => key.code == KeyCode::Up,
        "<Down>" => key.code == KeyCode::Down,
        "<Left>" => key.code == KeyCode::Left,
        "<Right>" => key.code == KeyCode::Right,
        "Space" | " " => key.code == KeyCode::Char(' '),
        s if s.len() == 1 => key.code == KeyCode::Char(s.chars().next().unwrap()),
        _ => false,
    }
}

pub fn binding_label(binding: &str) -> String {
    match binding {
        " " | "Space" => "Espacio".into(),
        s => s.to_string(),
    }
}
