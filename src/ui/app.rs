use std::time::{Duration, Instant};

use anyhow::Result;
use crossterm::event::{self, Event, KeyCode, KeyEvent};
use ratatui::{
    layout::{Constraint, Layout},
    style::Style,
    widgets::Block,
    Frame,
};

use crate::core::models::FileState;
use crate::core::state::{AppState, ToastKind, View};
use crate::ports::git_provider::GitProvider;
use crate::ui::{components, keybind, theme::Theme};

pub struct App {
    pub state: AppState,
    tick: usize,
    last_tick: Instant,
    repo_path: String,
    async_rx: Option<tokio::sync::mpsc::UnboundedReceiver<anyhow::Result<String>>>,
    hydrate_rx: Option<tokio::sync::oneshot::Receiver<anyhow::Result<(Vec<crate::core::models::FileEntry>, Vec<crate::core::models::FileEntry>, Option<crate::core::models::Branch>)>>>,
    last_content_height: usize,
}

impl App {
    pub fn new(state: AppState, repo_path: String) -> Self {
        Self {
            state,
            tick: 0,
            last_tick: Instant::now(),
            repo_path,
            async_rx: None,
            hydrate_rx: None,
            last_content_height: 20,
        }
    }

    pub fn run(
        &mut self,
        terminal: &mut ratatui::Terminal<ratatui::backend::CrosstermBackend<std::io::Stdout>>,
        git: &impl GitProvider,
    ) -> Result<()> {
        self.start_hydration();

        let tick_rate = Duration::from_millis(100);

        loop {
            self.check_toast_expiry();
            self.check_async_result(git)?;
            self.check_hydration()?;

            terminal.draw(|f| self.render(f))?;

            if self.state.should_quit {
                return Ok(());
            }

            if event::poll(tick_rate)? {
                if let Event::Key(key) = event::read()? {
                    self.handle_key(key, git)?;
                }
            }

            if self.last_tick.elapsed() >= tick_rate {
                self.tick = self.tick.wrapping_add(1);
                self.last_tick = Instant::now();
                self.decay_flash();
            }
        }
    }

    fn check_toast_expiry(&mut self) {
        if let Some(ref toast) = self.state.toast {
            if components::toast::should_dismiss(toast) {
                self.state.toast = None;
            }
        }
    }

    fn check_async_result(&mut self, git: &impl GitProvider) -> Result<()> {
        if let Some(rx) = &mut self.async_rx {
            if let Ok(result) = rx.try_recv() {
                self.async_rx = None;
                self.state.is_loading = false;
                match result {
                    Ok(msg) => self.state.notify(ToastKind::Success, msg),
                    Err(e) => self.state.notify(ToastKind::Error, format!("{}", e)),
                }
                self.refresh_state(git)?;
            }
        }
        Ok(())
    }

    fn refresh_state(&mut self, git: &impl GitProvider) -> Result<()> {
        let files = git.status().unwrap_or_default();
        let staged: Vec<_> = files
            .iter()
            .filter(|f| f.state == crate::core::models::FileState::Staged)
            .cloned()
            .collect();
        let rest: Vec<_> = files
            .into_iter()
            .filter(|f| f.state != crate::core::models::FileState::Staged)
            .collect();

        self.state.files = rest;
        self.state.staged = staged;
        self.state.branch = git.current_branch().ok();
        self.state.is_loading = false;

        let count = self.state.item_count();
        if count == 0 {
            self.state.selected_index = 0;
            self.state.scroll_offset = 0;
        } else if self.state.selected_index >= count {
            self.state.selected_index = count.saturating_sub(1);
        }

        Ok(())
    }

    fn render(&mut self, f: &mut Frame) {
        let theme = Theme::from_hex_strings(&self.state.config.theme);

        f.render_widget(
            Block::default().style(Style::default().bg(theme.bg_base)),
            f.area(),
        );

        let main_layout = Layout::default()
            .direction(ratatui::layout::Direction::Vertical)
            .constraints([Constraint::Min(1), Constraint::Length(1)])
            .split(f.area());

        let content_area = main_layout[0];
        let footer_area = main_layout[1];

        self.last_content_height = match self.state.view {
            View::Status => {
                let header_height = if self.state.branch.is_some() { 3 } else { 1 };
                content_area.height.saturating_sub(header_height).max(1) as usize
            }
            View::Log | View::Diff => {
                content_area.height.saturating_sub(2).max(1) as usize
            }
        };

        if !self.state.show_help {
            match self.state.view {
                View::Status => {
                    components::status_view::render_status(&self.state, f, content_area);
                }
                View::Log => {
                    components::log_view::render_log(&self.state, f, content_area);
                }
                View::Diff => {
                    components::diff_view::render_diff(&self.state, f, content_area);
                }
            }
        } else {
            components::help::render_help(&self.state, f, content_area);
        }

        if self.state.show_command_palette {
            components::command_palette::render_command_palette(&self.state, f, f.area());
        }

        if self.state.show_commit_input {
            components::commit_view::render_commit_input(&self.state, f, f.area());
        }

        if self.state.show_search {
            components::status_view::render_search_bar(&self.state, f, f.area());
        }

        components::spinner::render_spinner(&self.state, f, f.area(), self.tick);

        if self.state.toast.is_some() {
            components::toast::render_toast_raw(
                self.state.toast.as_ref().unwrap(),
                &theme,
                f,
                footer_area,
            );
        }

        components::footer::render_footer(&self.state, f, footer_area);
    }

    fn spawn_async(&mut self, op_name: &str, repo_path: String, f: fn(String) -> anyhow::Result<String>) {
        let (tx, rx) = tokio::sync::mpsc::unbounded_channel();
        let name = op_name.to_string();

        tokio::spawn(async move {
            let result = f(repo_path);
            let _ = tx.send(result);
        });

        self.async_rx = Some(rx);
        self.state.is_loading = true;
        self.state.loading_message = name;
    }

    fn handle_key(&mut self, key: KeyEvent, git: &impl GitProvider) -> Result<()> {
        if key.code == KeyCode::Char('c') && key.modifiers.contains(crossterm::event::KeyModifiers::CONTROL) {
            self.state.should_quit = true;
            return Ok(());
        }

        if self.state.show_help {
            if key.code == KeyCode::Esc || keybind::matches_binding(&key, &self.state.config.keybindings.help) {
                self.state.show_help = false;
            }
            return Ok(());
        }

        if self.state.show_command_palette {
            return self.handle_command_input(key);
        }

        if self.state.show_commit_input {
            return self.handle_commit_input(key, git);
        }

        if self.state.show_search {
            return self.handle_search_input(key);
        }

        if keybind::matches_binding(&key, &self.state.config.keybindings.help) {
            self.state.show_help = !self.state.show_help;
            return Ok(());
        }

        if keybind::matches_binding(&key, &self.state.config.keybindings.command_palette) {
            self.state.show_command_palette = true;
            self.state.command_input.clear();
            return Ok(());
        }

        if key.code == KeyCode::Tab {
            self.state.show_search = false;
            self.state.search_query.clear();
            match self.state.view {
                View::Status => self.state.view = View::Log,
                View::Log => self.state.view = View::Diff,
                View::Diff => self.state.view = View::Status,
            }
            self.state.selected_index = 0;
            self.state.scroll_offset = 0;
            return Ok(());
        }

        match self.state.view {
            View::Status => self.handle_status_key(key, git)?,
            View::Log => self.handle_log_key(key, git)?,
            View::Diff => self.handle_diff_key(key)?,
        }

        Ok(())
    }

    fn handle_search_input(&mut self, key: KeyEvent) -> Result<()> {
        match key.code {
            KeyCode::Esc => {
                self.state.show_search = false;
                self.state.search_query.clear();
            }
            KeyCode::Enter => {
                self.state.show_search = false;
            }
            KeyCode::Char(c) => {
                self.state.search_query.push(c);
            }
            KeyCode::Backspace => {
                self.state.search_query.pop();
            }
            _ => {}
        }
        Ok(())
    }

    fn decay_flash(&mut self) {
        let cutoff = Instant::now() - Duration::from_millis(800);
        self.state.flash_files.retain(|(_, t)| *t > cutoff);
    }

    fn start_hydration(&mut self) {
        let (tx, rx) = tokio::sync::oneshot::channel();
        let path = self.repo_path.clone();

        tokio::task::spawn_blocking(move || {
            let result = (|| -> anyhow::Result<_> {
                let adapter = crate::adapters::git2_adapter::Git2Adapter::new(&path)?;
                let files = adapter.status()?;
                let staged: Vec<_> = files
                    .iter()
                    .filter(|f| f.state == crate::core::models::FileState::Staged)
                    .cloned()
                    .collect();
                let rest: Vec<_> = files
                    .into_iter()
                    .filter(|f| f.state != crate::core::models::FileState::Staged)
                    .collect();
                let branch = adapter.current_branch().ok();
                Ok((rest, staged, branch))
            })();
            let _ = tx.send(result);
        });

        self.hydrate_rx = Some(rx);
    }

    fn check_hydration(&mut self) -> Result<()> {
        if let Some(rx) = &mut self.hydrate_rx {
            if let Ok(result) = rx.try_recv() {
                self.hydrate_rx = None;
                match result {
                    Ok((files, staged, branch)) => {
                        self.state.files = files;
                        self.state.staged = staged;
                        self.state.branch = branch;
                        self.state.is_loading = false;

                        let count = self.state.item_count();
                        if count == 0 {
                            self.state.selected_index = 0;
                            self.state.scroll_offset = 0;
                        } else if self.state.selected_index >= count {
                            self.state.selected_index = count.saturating_sub(1);
                        }
                    }
                    Err(e) => {
                        self.state.notify(ToastKind::Error, format!("Error al cargar repositorio: {}", e));
                        self.state.is_loading = false;
                    }
                }
            }
        }
        Ok(())
    }

    fn handle_status_key(&mut self, key: KeyEvent, git: &impl GitProvider) -> Result<()> {
        let visible = self.last_content_height;
        let kb = &self.state.config.keybindings;

        if keybind::matches_binding(&key, &kb.down) || key.code == KeyCode::Down {
            self.state.select_next(visible);
            return Ok(());
        }
        if keybind::matches_binding(&key, &kb.up) || key.code == KeyCode::Up {
            self.state.select_prev(visible);
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.toggle_stage) {
            if let Some(file) = self.state.selected_file() {
                let path = file.path.clone();
                let was_staged = matches!(file.state, FileState::Staged);
                let now = Instant::now();
                if was_staged {
                    git.unstage(&path)?;
                    self.state.notify(ToastKind::Success, format!("Unstage: {}", path));
                } else {
                    git.stage(&path)?;
                    self.state.notify(ToastKind::Success, format!("Stage: {}", path));
                }
                self.state.flash_files.retain(|(p, _)| p != &path);
                self.state.flash_files.push((path, now));
                self.refresh_state(git)?;
            }
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.stage) {
            if let Some(file) = self.state.selected_file() {
                let path = file.path.clone();
                if !matches!(file.state, FileState::Staged) {
                    git.stage(&path)?;
                    self.state.notify(ToastKind::Success, format!("Stage: {}", path));
                    let now = Instant::now();
                    self.state.flash_files.retain(|(p, _)| p != &path);
                    self.state.flash_files.push((path, now));
                    self.refresh_state(git)?;
                }
            }
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.unstage) {
            if let Some(file) = self.state.selected_file() {
                let path = file.path.clone();
                git.unstage(&path)?;
                self.state.notify(ToastKind::Success, format!("Unstage: {}", path));
                let now = Instant::now();
                self.state.flash_files.retain(|(p, _)| p != &path);
                self.state.flash_files.push((path, now));
                self.refresh_state(git)?;
            }
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.commit) {
            self.state.show_commit_input = true;
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.log_view) {
            self.state.commits = git.log(50).unwrap_or_default();
            self.state.view = View::Log;
            self.state.selected_index = 0;
            self.state.scroll_offset = 0;
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.diff_view) || key.code == KeyCode::Enter {
            if let Some(file) = self.state.selected_file() {
                let path = file.path.clone();
                self.state.diff_content = git.diff_file(&path).ok();
                self.state.view = View::Diff;
                self.state.scroll_offset = 0;
            }
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.search) {
            self.state.show_search = true;
            self.state.search_query.clear();
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.push) {
            self.spawn_async("Haciendo push...", self.repo_path.clone(), |path| {
                let adapter = crate::adapters::git2_adapter::Git2Adapter::new(&path)?;
                adapter.push()?;
                Ok("Push completado".into())
            });
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.pull) {
            self.spawn_async("Haciendo pull...", self.repo_path.clone(), |path| {
                let adapter = crate::adapters::git2_adapter::Git2Adapter::new(&path)?;
                adapter.pull()?;
                Ok("Pull completado".into())
            });
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.quit) {
            self.state.should_quit = true;
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.command_palette) {
            self.state.show_command_palette = true;
            self.state.command_input.clear();
        }

        Ok(())
    }

    fn handle_log_key(&mut self, key: KeyEvent, git: &impl GitProvider) -> Result<()> {
        let visible = self.last_content_height;
        let kb = &self.state.config.keybindings;

        if keybind::matches_binding(&key, &kb.down) || key.code == KeyCode::Down {
            self.state.select_next(visible);
            return Ok(());
        }
        if keybind::matches_binding(&key, &kb.up) || key.code == KeyCode::Up {
            self.state.select_prev(visible);
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.diff_view) || key.code == KeyCode::Enter {
            if let Some(commit) = self.state.selected_commit() {
                self.state.diff_content = Some(format!(
                    "commit {}\nAuthor: {}\n\n{}",
                    commit.hash, commit.author, commit.message
                ));
                self.state.view = View::Diff;
                self.state.scroll_offset = 0;
            }
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.back)
            || key.code == KeyCode::Esc
        {
            self.state.view = View::Status;
            self.state.selected_index = 0;
            self.state.scroll_offset = 0;
            self.refresh_state(git)?;
        }

        Ok(())
    }

    fn handle_diff_key(&mut self, key: KeyEvent) -> Result<()> {
        let visible = self.last_content_height;
        let kb = &self.state.config.keybindings;

        if keybind::matches_binding(&key, &kb.down) || key.code == KeyCode::Down {
            self.state.select_next(visible);
            return Ok(());
        }
        if keybind::matches_binding(&key, &kb.up) || key.code == KeyCode::Up {
            self.state.select_prev(visible);
            return Ok(());
        }

        if keybind::matches_binding(&key, &kb.back)
            || key.code == KeyCode::Esc
            || keybind::matches_binding(&key, &kb.quit)
        {
            self.state.view = View::Status;
            self.state.selected_index = 0;
            self.state.scroll_offset = 0;
        }

        Ok(())
    }

    fn handle_command_input(&mut self, key: KeyEvent) -> Result<()> {
        match key.code {
            KeyCode::Esc => {
                self.state.clear_command();
            }
            KeyCode::Enter => {
                let cmd = self.state.command_input.clone();
                self.state.clear_command();
                self.execute_command(&cmd)?;
            }
            KeyCode::Char(c) => {
                self.state.command_input.push(c);
            }
            KeyCode::Backspace => {
                self.state.command_input.pop();
            }
            _ => {}
        }
        Ok(())
    }

    fn handle_commit_input(&mut self, key: KeyEvent, git: &impl GitProvider) -> Result<()> {
        match key.code {
            KeyCode::Esc => {
                self.state.show_commit_input = false;
                self.state.commit_message.clear();
            }
            KeyCode::Enter => {
                let msg = self.state.commit_message.clone();
                if !msg.trim().is_empty() {
                    git.commit(msg.trim())?;
                    self.state.notify(ToastKind::Success, "Commit creado");
                    self.refresh_state(git)?;
                }
                self.state.show_commit_input = false;
                self.state.commit_message.clear();
            }
            KeyCode::Char(c) => {
                self.state.commit_message.push(c);
            }
            KeyCode::Backspace => {
                self.state.commit_message.pop();
            }
            _ => {}
        }
        Ok(())
    }

    fn execute_command(&mut self, cmd: &str) -> Result<()> {
        match cmd.trim() {
            "q" | "quit" => self.state.should_quit = true,
            "status" => {
                self.state.view = View::Status;
                self.state.selected_index = 0;
                self.state.scroll_offset = 0;
            }
            "log" => {
                self.state.view = View::Log;
                self.state.selected_index = 0;
                self.state.scroll_offset = 0;
            }
            "commit" => {
                self.state.show_commit_input = true;
            }
            "push" => {
                self.spawn_async("Haciendo push...", self.repo_path.clone(), |path| {
                    let adapter = crate::adapters::git2_adapter::Git2Adapter::new(&path)?;
                    adapter.push()?;
                    Ok("Push completado".into())
                });
            }
            "pull" => {
                self.spawn_async("Haciendo pull...", self.repo_path.clone(), |path| {
                    let adapter = crate::adapters::git2_adapter::Git2Adapter::new(&path)?;
                    adapter.pull()?;
                    Ok("Pull completado".into())
                });
            }
            _ => {
                self.state.notify(ToastKind::Info, format!("Comando: {}", cmd));
            }
        }
        Ok(())
    }
}
