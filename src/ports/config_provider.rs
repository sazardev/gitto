use anyhow::Result;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct ThemeConfig {
    pub bg_base: String,
    pub accent_primary: String,
    pub success: String,
    pub danger: String,
    pub muted: String,
}

impl Default for ThemeConfig {
    fn default() -> Self {
        Self {
            bg_base: String::from("#1E1E2E"),
            accent_primary: String::from("#CBA6F7"),
            success: String::from("#A6E3A1"),
            danger: String::from("#F38BA8"),
            muted: String::from("#6C7086"),
        }
    }
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct Keybindings {
    #[serde(default)]
    pub up: String,
    #[serde(default)]
    pub down: String,
    #[serde(default)]
    pub toggle_stage: String,
    #[serde(default)]
    pub stage: String,
    #[serde(default)]
    pub unstage: String,
    #[serde(default)]
    pub commit: String,
    #[serde(default)]
    pub push: String,
    #[serde(default)]
    pub pull: String,
    #[serde(default)]
    pub log_view: String,
    #[serde(default)]
    pub diff_view: String,
    #[serde(default)]
    pub command_palette: String,
    #[serde(default)]
    pub search: String,
    #[serde(default)]
    pub help: String,
    #[serde(default)]
    pub back: String,
    #[serde(default)]
    pub move_left: String,
    #[serde(default)]
    pub move_right: String,
    #[serde(default)]
    pub quit: String,
    #[serde(default)]
    pub toggle_split_horizontal: String,
    #[serde(default)]
    pub toggle_split_vertical: String,
    #[serde(default)]
    pub toggle_zen: String,
    #[serde(default)]
    pub switch_panel: String,
}

impl Default for Keybindings {
    fn default() -> Self {
        Self {
            up: String::from("k"),
            down: String::from("j"),
            toggle_stage: String::from(" "),
            stage: String::from("s"),
            unstage: String::from("u"),
            commit: String::from("c"),
            push: String::from("P"),
            pull: String::from("p"),
            log_view: String::from("l"),
            diff_view: String::from("<Enter>"),
            command_palette: String::from(":"),
            search: String::from("\\"),
            help: String::from("?"),
            back: String::from("<Esc>"),
            move_left: String::from("h"),
            move_right: String::from("l"),
            quit: String::from("q"),
            toggle_split_horizontal: String::from("Ctrl+H"),
            toggle_split_vertical: String::from("Ctrl+V"),
            toggle_zen: String::from("Ctrl+Z"),
            switch_panel: String::from("Ctrl+W"),
        }
    }
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct AppConfig {
    #[serde(default)]
    pub theme: ThemeConfig,
    #[serde(default)]
    pub keybindings: Keybindings,
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            theme: ThemeConfig::default(),
            keybindings: Keybindings::default(),
        }
    }
}

pub trait ConfigProvider {
    fn load() -> Result<AppConfig>;
}
