use std::time::Instant;
use crate::ports::config_provider::AppConfig;
use super::models::*;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum View {
    Status,
    Log,
    Diff,
}

#[derive(Debug, Clone)]
pub struct ToastMessage {
    pub kind: ToastKind,
    pub text: String,
    pub created_at: Instant,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ToastKind {
    Success,
    Error,
    Info,
}

pub struct AppState {
    pub view: View,
    pub files: Vec<FileEntry>,
    pub staged: Vec<FileEntry>,
    pub commits: Vec<Commit>,
    pub branch: Option<Branch>,
    pub selected_index: usize,
    pub scroll_offset: usize,
    pub diff_content: Option<String>,
    pub command_input: String,
    pub show_command_palette: bool,
    pub command_matches: Vec<String>,
    pub show_commit_input: bool,
    pub commit_message: String,
    pub toast: Option<ToastMessage>,
    pub is_loading: bool,
    pub loading_message: String,
    pub config: AppConfig,
    pub should_quit: bool,
}

impl AppState {
    pub fn new(config: AppConfig) -> Self {
        Self {
            view: View::Status,
            files: Vec::new(),
            staged: Vec::new(),
            commits: Vec::new(),
            branch: None,
            selected_index: 0,
            scroll_offset: 0,
            diff_content: None,
            command_input: String::new(),
            show_command_palette: false,
            command_matches: Vec::new(),
            show_commit_input: false,
            commit_message: String::new(),
            toast: None,
            is_loading: true,
            loading_message: String::from("Cargando repositorio..."),
            config,
            should_quit: false,
        }
    }

    pub fn all_files(&self) -> Vec<&FileEntry> {
        self.staged.iter().chain(self.files.iter()).collect()
    }

    pub fn staged_count(&self) -> usize {
        self.staged.len()
    }

    pub fn unstaged_or_untracked_count(&self) -> usize {
        self.files.len()
    }

    pub fn selected_file(&self) -> Option<&FileEntry> {
        self.all_files().get(self.selected_index).copied()
    }

    pub fn selected_commit(&self) -> Option<&Commit> {
        self.commits.get(self.selected_index)
    }

    pub fn select_next(&mut self, visible_rows: usize) {
        let len = self.item_count();
        if len == 0 {
            return;
        }
        if self.selected_index < len.saturating_sub(1) {
            self.selected_index += 1;
        }
        self.clamp_scroll(visible_rows);
    }

    pub fn select_prev(&mut self, visible_rows: usize) {
        if self.selected_index > 0 {
            self.selected_index -= 1;
        }
        self.clamp_scroll(visible_rows);
    }

    pub fn clamp_scroll(&mut self, visible_rows: usize) {
        let max_offset = self.item_count().saturating_sub(visible_rows);
        if self.selected_index < self.scroll_offset {
            self.scroll_offset = self.selected_index;
        } else if self.selected_index >= self.scroll_offset.saturating_add(visible_rows) {
            self.scroll_offset = self.selected_index.saturating_sub(visible_rows.saturating_sub(1));
        }
        if self.scroll_offset > max_offset {
            self.scroll_offset = max_offset;
        }
    }

    pub fn item_count(&self) -> usize {
        match self.view {
            View::Status => self.all_files().len(),
            View::Log => self.commits.len(),
            View::Diff => 1,
        }
    }

    pub fn clear_command(&mut self) {
        self.show_command_palette = false;
        self.command_input.clear();
        self.command_matches.clear();
    }

    pub fn notify(&mut self, kind: ToastKind, text: impl Into<String>) {
        self.toast = Some(ToastMessage {
            kind,
            text: text.into(),
            created_at: Instant::now(),
        });
    }
}
