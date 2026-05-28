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
    pub stage: String,
    pub unstage: String,
    pub commit: String,
    pub push: String,
    pub pull: String,
    pub log_view: String,
    pub diff_view: String,
    pub command_palette: String,
    pub quit: String,
}

impl Default for Keybindings {
    fn default() -> Self {
        Self {
            stage: String::from("s"),
            unstage: String::from("u"),
            commit: String::from("c"),
            push: String::from("P"),
            pull: String::from("p"),
            log_view: String::from("l"),
            diff_view: String::from("<Enter>"),
            command_palette: String::from(":"),
            quit: String::from("q"),
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
