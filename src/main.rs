mod adapters;
mod core;
mod ports;
mod ui;

use ports::config_provider::ConfigProvider;
use ports::git_provider::GitProvider;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let config = adapters::fs_config::FsConfig::load()?;
    let state = core::state::AppState::new(config);

    let cwd = std::env::current_dir()
        .map(|p| p.to_string_lossy().to_string())
        .unwrap_or_else(|_| String::from("."));

    let repo_path = adapters::git2_adapter::Git2Adapter::discover(&cwd)?;
    let git = adapters::git2_adapter::Git2Adapter::new(&repo_path)?;

    crossterm::terminal::enable_raw_mode()?;
    let mut stdout = std::io::stdout();
    crossterm::execute!(
        stdout,
        crossterm::terminal::EnterAlternateScreen,
        crossterm::cursor::Hide
    )?;

    let backend = ratatui::backend::CrosstermBackend::new(stdout);
    let mut terminal = ratatui::Terminal::new(backend)?;

    let result = {
        let mut app = ui::app::App::new(state, repo_path);
        app.run(&mut terminal, &git)
    };

    crossterm::terminal::disable_raw_mode()?;
    crossterm::execute!(
        terminal.backend_mut(),
        crossterm::terminal::LeaveAlternateScreen,
        crossterm::cursor::Show
    )?;

    result
}
