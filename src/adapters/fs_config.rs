use anyhow::Result;
use crate::ports::config_provider::{AppConfig, ConfigProvider};

pub struct FsConfig;

impl ConfigProvider for FsConfig {
    fn load() -> Result<AppConfig> {
        let config_dir = dirs::config_dir()
            .unwrap_or_else(|| std::path::PathBuf::from("."))
            .join("gitto");

        let config_path = config_dir.join("config.toml");

        if config_path.exists() {
            let content = std::fs::read_to_string(&config_path)?;
            let config: AppConfig = toml::from_str(&content)?;
            Ok(config)
        } else {
            let default = AppConfig::default();
            if let Ok(toml_str) = toml::to_string_pretty(&default) {
                if std::fs::create_dir_all(&config_dir).is_ok() {
                    let _ = std::fs::write(&config_path, toml_str);
                }
            }
            Ok(default)
        }
    }
}
