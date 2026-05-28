

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum FileState {
    Unstaged,
    Staged,
    Untracked,
    Renamed,
    Conflicted,
}

#[derive(Debug, Clone)]
pub struct FileEntry {
    pub path: String,
    pub state: FileState,
    pub old_path: Option<String>,
}

#[derive(Debug, Clone)]
pub struct Commit {
    pub hash: String,
    pub message: String,
    pub author: String,
    pub timestamp: i64,
}

#[derive(Debug, Clone)]
pub struct Branch {
    pub name: String,
    pub upstream: Option<String>,
    pub ahead: usize,
    pub behind: usize,
}
