use std::time::{Duration, Instant};

use anyhow::Result;
use crossterm::event::{self, Event, KeyCode, KeyEvent, KeyModifiers};
use ratatui::{
    layout::{Constraint, Layout},
    style::Style,
    widgets::Block,
    Frame,
};

use crate::core::state::{AppState, ToastKind, View};
use crate::ports::git_provider::GitProvider;
use crate::ui::components;
use crate::ui::theme::Theme;

pub struct App {
    pub state: AppState,
    tick: usize,
    last_tick: Instant,
}

impl App {
    pub fn new(state: AppState) -> Self {
        Self {
            state,
            tick: 0,
            last_tick: Instant::now(),
        }
    }

    pub fn run(
        &mut self,
        terminal: &mut ratatui::Terminal<ratatui::backend::CrosstermBackend<std::io::Stdout>>,
        git: &impl GitProvider,
    ) -> Result<()> {
        self.refresh_state(git)?;

        let tick_rate = Duration::from_millis(100);

        loop {
            self.check_toast_expiry();

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

    fn render(&self, f: &mut Frame) {
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

        if self.state.show_command_palette {
            components::command_palette::render_command_palette(&self.state, f, f.area());
        }

        if self.state.show_commit_input {
            components::commit_view::render_commit_input(&self.state, f, f.area());
        }

        components::spinner::render_spinner(&self.state, f, f.area(), self.tick);

        if self.state.toast.is_some() {
            let toast_area = footer_area;
            components::toast::render_toast_raw(
                self.state.toast.as_ref().unwrap(),
                &theme,
                f,
                toast_area,
            );
        }

        components::footer::render_footer(&self.state, f, footer_area);
    }

    fn handle_key(&mut self, key: KeyEvent, git: &impl GitProvider) -> Result<()> {
        if key.modifiers.contains(KeyModifiers::CONTROL) && key.code == KeyCode::Char('q') {
            self.state.should_quit = true;
            return Ok(());
        }

        if self.state.show_command_palette {
            return self.handle_command_input(key);
        }

        if self.state.show_commit_input {
            return self.handle_commit_input(key, git);
        }

        match self.state.view {
            View::Status => self.handle_status_key(key, git)?,
            View::Log => self.handle_log_key(key, git)?,
            View::Diff => self.handle_diff_key(key)?,
        }

        Ok(())
    }

    fn visible_rows_for_view(&self) -> usize {
        20
    }

    fn handle_status_key(&mut self, key: KeyEvent, git: &impl GitProvider) -> Result<()> {
        let visible = self.visible_rows_for_view();
        match key.code {
            KeyCode::Char('j') | KeyCode::Down => self.state.select_next(visible),
            KeyCode::Char('k') | KeyCode::Up => self.state.select_prev(visible),
            KeyCode::Char('s') => {
                if let Some(file) = self.state.selected_file() {
                    let path = file.path.clone();
                    git.stage(&path)?;
                    self.state.notify(ToastKind::Success, format!("Stage: {}", path));
                    self.refresh_state(git)?;
                }
            }
            KeyCode::Char('u') => {
                if let Some(file) = self.state.selected_file() {
                    let path = file.path.clone();
                    git.unstage(&path)?;
                    self.state.notify(ToastKind::Success, format!("Unstage: {}", path));
                    self.refresh_state(git)?;
                }
            }
            KeyCode::Char('c') => {
                self.state.show_commit_input = true;
            }
            KeyCode::Char('l') => {
                self.state.commits = git.log(50).unwrap_or_default();
                self.state.view = View::Log;
                self.state.selected_index = 0;
                self.state.scroll_offset = 0;
            }
            KeyCode::Char(':') => {
                self.state.show_command_palette = true;
                self.state.command_input.clear();
            }
            KeyCode::Enter => {
                if let Some(file) = self.state.selected_file() {
                    let path = file.path.clone();
                    self.state.diff_content = git.diff_file(&path).ok();
                    self.state.view = View::Diff;
                    self.state.scroll_offset = 0;
                }
            }
            KeyCode::Char('P') => {
                self.state.is_loading = true;
                self.state.loading_message = String::from("Haciendo push...");
                match git.push() {
                    Ok(()) => {
                        self.state.notify(ToastKind::Success, String::from("Push completado"));
                    }
                    Err(e) => {
                        self.state
                            .notify(ToastKind::Error, format!("Push fallido: {}", e));
                    }
                }
                self.state.is_loading = false;
                self.refresh_state(git)?;
            }
            KeyCode::Char('p') => {
                self.state.is_loading = true;
                self.state.loading_message = String::from("Haciendo pull...");
                match git.pull() {
                    Ok(()) => {
                        self.state.notify(ToastKind::Success, String::from("Pull completado"));
                    }
                    Err(e) => {
                        self.state
                            .notify(ToastKind::Error, format!("Pull fallido: {}", e));
                    }
                }
                self.state.is_loading = false;
                self.refresh_state(git)?;
            }
            KeyCode::Char('q') => self.state.should_quit = true,
            _ => {}
        }
        Ok(())
    }

    fn handle_log_key(&mut self, key: KeyEvent, git: &impl GitProvider) -> Result<()> {
        let visible = self.visible_rows_for_view();
        match key.code {
            KeyCode::Char('j') | KeyCode::Down => self.state.select_next(visible),
            KeyCode::Char('k') | KeyCode::Up => self.state.select_prev(visible),
            KeyCode::Enter => {
                if let Some(commit) = self.state.selected_commit() {
                    self.state.diff_content = Some(format!(
                        "commit {}\nAuthor: {}\n\n{}",
                        commit.hash, commit.author, commit.message
                    ));
                    self.state.view = View::Diff;
                    self.state.scroll_offset = 0;
                }
            }
            KeyCode::Esc | KeyCode::Char('q') | KeyCode::Char('l') => {
                self.state.view = View::Status;
                self.state.selected_index = 0;
                self.state.scroll_offset = 0;
                self.refresh_state(git)?;
            }
            _ => {}
        }
        Ok(())
    }

    fn handle_diff_key(&mut self, key: KeyEvent) -> Result<()> {
        let visible = self.visible_rows_for_view();
        match key.code {
            KeyCode::Char('j') | KeyCode::Down => self.state.select_next(visible),
            KeyCode::Char('k') | KeyCode::Up => self.state.select_prev(visible),
            KeyCode::Esc | KeyCode::Char('q') => {
                self.state.view = View::Status;
                self.state.selected_index = 0;
                self.state.scroll_offset = 0;
            }
            _ => {}
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
                    self.state.notify(ToastKind::Success, format!("Commit creado"));
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
            _ => {
                self.state
                    .notify(ToastKind::Info, format!("Comando desconocido: {}", cmd));
            }
        }
        Ok(())
    }
}
