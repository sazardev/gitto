use anyhow::Result;
use crate::ports::config_provider::{AppConfig, ConfigProvider};

pub struct FsConfig;

impl ConfigProvider for FsConfig {
    fn load() -> Result<AppConfig> {
        let config_path = dirs::config_dir()
            .unwrap_or_else(|| std::path::PathBuf::from("."))
            .join("gitto")
            .join("config.toml");

        if config_path.exists() {
            let content = std::fs::read_to_string(&config_path)?;
            let config: AppConfig = toml::from_str(&content)?;
            Ok(config)
        } else {
            Ok(AppConfig::default())
        }
    }
}
