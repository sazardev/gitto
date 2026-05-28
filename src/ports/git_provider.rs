use anyhow::Result;
use crate::core::models::*;

pub trait GitProvider {
    fn discover(path: &str) -> Result<String>;
    fn status(&self) -> Result<Vec<FileEntry>>;
    fn stage(&self, path: &str) -> Result<()>;
    fn unstage(&self, path: &str) -> Result<()>;
    fn commit(&self, message: &str) -> Result<String>;
    fn push(&self) -> Result<()>;
    fn pull(&self) -> Result<()>;
    fn log(&self, count: usize) -> Result<Vec<Commit>>;
    fn diff_file(&self, path: &str) -> Result<String>;
    fn current_branch(&self) -> Result<Branch>;
}
