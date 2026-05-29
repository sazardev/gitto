use std::time::Instant;
use crate::ports::config_provider::AppConfig;
use super::models::FileEntry;
use super::models::FileState;
use super::models::Commit;
use super::models::Branch;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum View {
    Status,
    Log,
    Diff,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Layout {
    Zen,
    SplitHorizontal,
    SplitVertical,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Panel {
    Left,
    Right,
    Top,
    Bottom,
    Main,
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
    pub show_help: bool,
    pub is_loading: bool,
    pub loading_message: String,
    pub config: AppConfig,
    pub should_quit: bool,
    pub show_search: bool,
    pub search_query: String,
    pub flash_files: Vec<(String, Instant)>,
    pub layout: Layout,
    pub active_panel: Panel,
    pub secondary_view: Option<View>,
    pub staging_area: Vec<StagedFile>,
    pub unstaged_area: Vec<StagedFile>,
}

#[derive(Debug, Clone)]
pub struct StagedFile {
    pub path: String,
    pub hunks: Vec<DiffHunk>,
    pub staged: bool,
}

#[derive(Debug, Clone)]
pub struct DiffHunk {
    pub header: String,
    pub lines: Vec<DiffLine>,
}

#[derive(Debug, Clone)]
pub struct DiffLine {
    pub content: String,
    pub change_type: ChangeType,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ChangeType {
    Added,
    Removed,
    Context,
    Header,
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
            show_help: false,
            is_loading: true,
            loading_message: String::from("Cargando repositorio..."),
            config,
            should_quit: false,
            show_search: false,
            search_query: String::new(),
            flash_files: Vec::new(),
            layout: Layout::Zen,
            active_panel: Panel::Main,
            secondary_view: None,
            staging_area: Vec::new(),
            unstaged_area: Vec::new(),
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

    pub fn visual_map_for(&self, files: &[&FileEntry]) -> Vec<Option<usize>> {
        let mut map = Vec::new();
        let mut last_section: Option<&str> = None;

        for (i, file) in files.iter().enumerate() {
            let section = section_name(file);
            if last_section != Some(section) {
                if last_section.is_some() {
                    map.push(None);
                }
                map.push(None);
                last_section = Some(section);
            }
            map.push(Some(i));
        }
        map
    }

    pub fn visual_map(&self) -> Vec<Option<usize>> {
        let files = self.all_files();
        self.visual_map_for(&files)
    }

    pub fn visual_index(&self) -> usize {
        let map = self.visual_map();
        for (visual, item) in map.iter().enumerate() {
            if *item == Some(self.selected_index) {
                return visual;
            }
        }
        0
    }

    pub fn clamp_scroll(&mut self, visible_rows: usize) {
        let map = self.visual_map();
        let visual_count = map.len();
        if visual_count == 0 {
            return;
        }

        let visual_idx = self.visual_index();
        let visual_max = visual_count.saturating_sub(1);
        let max_offset = visual_max.saturating_sub(visible_rows.saturating_sub(1));

        if visual_idx < self.scroll_offset {
            self.scroll_offset = visual_idx;
        } else if visual_idx >= self.scroll_offset.saturating_add(visible_rows) {
            self.scroll_offset = visual_idx.saturating_sub(visible_rows.saturating_sub(1));
        }

        if self.scroll_offset > max_offset {
            self.scroll_offset = max_offset;
        }
    }

    pub fn visual_count(&self) -> usize {
        self.visual_map().len()
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

fn section_name(file: &FileEntry) -> &'static str {
    match file.state {
        FileState::Staged => "Cambios Staged",
        FileState::Unstaged => "Cambios sin Stagear",
        FileState::Untracked => "Sin Seguimiento",
        FileState::Renamed => "Renombrados",
        FileState::Conflicted => "Conflictos",
    }
}
