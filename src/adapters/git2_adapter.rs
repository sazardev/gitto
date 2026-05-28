use anyhow::{Context, Result};
use git2::{Repository, StatusOptions, StatusShow};
use crate::core::models::*;
use crate::ports::git_provider::GitProvider;

pub struct Git2Adapter {
    repo: Repository,
}

impl GitProvider for Git2Adapter {
    fn discover(path: &str) -> Result<String> {
        let repo = Repository::discover(path)
            .context("No se encontro un repositorio Git")?;
        let workdir = repo.workdir()
            .context("El repositorio no tiene directorio de trabajo")?;
        Ok(workdir.to_string_lossy().to_string())
    }

    fn status(&self) -> Result<Vec<FileEntry>> {
        let mut opts = StatusOptions::new();
        opts.show(StatusShow::IndexAndWorkdir)
            .include_untracked(true)
            .renames_head_to_index(true);

        let statuses = self.repo.statuses(Some(&mut opts))
            .context("Error al leer el estado del repositorio")?;

        let mut files = Vec::new();

        for entry in statuses.iter() {
            let path = entry.path().unwrap_or("").to_string();
            let status = entry.status();

            let state = if status.is_index_new()
                || status.is_index_modified()
                || status.is_index_deleted()
                || status.is_index_renamed()
                || status.is_index_typechange()
            {
                FileState::Staged
            } else if status.is_wt_new() {
                FileState::Untracked
            } else if status.is_conflicted() {
                FileState::Conflicted
            } else if status.is_wt_renamed() {
                FileState::Renamed
            } else {
                FileState::Unstaged
            };

            let old_path = if status.is_wt_renamed() {
                entry.index_to_workdir()
                    .and_then(|d| d.old_file().path())
                    .map(|p| p.to_string_lossy().to_string())
            } else if status.is_index_renamed() {
                entry.head_to_index()
                    .and_then(|d| d.old_file().path())
                    .map(|p| p.to_string_lossy().to_string())
            } else {
                None
            };

            files.push(FileEntry {
                path,
                state,
                old_path,
            });
        }

        Ok(files)
    }

    fn stage(&self, path: &str) -> Result<()> {
        let mut index = self.repo.index().context("Error al abrir el indice")?;
        index.add_path(std::path::Path::new(path))?;
        index.write()?;
        Ok(())
    }

    fn unstage(&self, path: &str) -> Result<()> {
        let head = self.repo.head()
            .context("Error al obtener HEAD")?;
        let head_tree = head.peel_to_tree()?;
        self.repo.reset_default(
            Some(head_tree.as_object()),
            &[path],
        )?;
        Ok(())
    }

    fn commit(&self, message: &str) -> Result<String> {
        let sig = self.repo.signature()
            .context("Error al obtener la firma de Git")?;
        let mut index = self.repo.index()
            .context("Error al abrir el indice")?;
        let tree_oid = index.write_tree()
            .context("Error al escribir el arbol")?;
        let tree = self.repo.find_tree(tree_oid)
            .context("Error al encontrar el arbol")?;

        let head = self.repo.head().ok();
        let head_commit = head.as_ref().and_then(|h| h.peel_to_commit().ok());
        let parent_refs: Vec<&git2::Commit> = head_commit.iter().collect();

        let oid = self.repo.commit(
            Some("HEAD"),
            &sig,
            &sig,
            message,
            &tree,
            &parent_refs,
        )?;

        Ok(oid.to_string())
    }

    fn push(&self) -> Result<()> {
        let mut remote = self.repo.find_remote("origin")
            .context("No se encontro el remoto 'origin'")?;

        let head = self.repo.head()?;
        let branch_name = head.shorthand()
            .context("No se pudo obtener el nombre de la rama")?;

        let refspec = format!("refs/heads/{}:refs/heads/{}", branch_name, branch_name);
        let callbacks = git2::RemoteCallbacks::new();
        let mut opts = git2::PushOptions::new();
        opts.remote_callbacks(callbacks);

        remote.push(&[&refspec], Some(&mut opts))
            .context("Error al hacer push")?;

        Ok(())
    }

    fn pull(&self) -> Result<()> {
        let mut remote = self.repo.find_remote("origin")
            .context("No se encontro el remoto 'origin'")?;

        let head = self.repo.head()?;
        let branch_name = head.shorthand()
            .context("No se pudo obtener el nombre de la rama")?;

        let refspec = format!("refs/heads/{}:refs/heads/{}", branch_name, branch_name);
        remote.fetch(&[&refspec], None, None)
            .context("Error al hacer fetch del remoto")?;

        let fetch_head = self.repo.find_reference("FETCH_HEAD")?;
        let fetch_commit = fetch_head.peel_to_commit()?;
        let annotated = self.repo.reference_to_annotated_commit(&fetch_head)?;

        let analysis = self.repo.merge_analysis(&[&annotated])?;
        if analysis.0.is_fast_forward() {
            let mut head_ref = self.repo.find_reference("HEAD")?;
            let msg = format!("Fast-forward: merging origin/{}", branch_name);
            head_ref.set_target(fetch_commit.id(), &msg)?;
            self.repo.set_head(head_ref.name().unwrap_or("HEAD"))?;
            self.repo.checkout_head(Some(git2::build::CheckoutBuilder::default().force()))?;
        } else if analysis.0.is_normal() {
            self.repo.merge(&[&annotated], None, None)?;
            let mut index = self.repo.index()?;
            let tree_oid = index.write_tree_to(&self.repo)?;
            let tree = self.repo.find_tree(tree_oid)?;
            let sig = self.repo.signature()?;
            let head_commit = self.repo.head()?.peel_to_commit()?;
            self.repo.commit(
                Some("HEAD"),
                &sig,
                &sig,
                &format!("Merge branch 'origin/{}'", branch_name),
                &tree,
                &[&head_commit, &fetch_commit],
            )?;
            self.repo.checkout_head(Some(git2::build::CheckoutBuilder::default().force()))?;
        }

        Ok(())
    }

    fn log(&self, count: usize) -> Result<Vec<Commit>> {
        let mut revwalk = self.repo.revwalk()
            .context("Error al crear el revwalk")?;
        revwalk.push_head()?;

        let mut commits = Vec::new();

        for oid in revwalk.take(count) {
            let oid = oid.context("Error al recorrer commits")?;
            let commit = self.repo.find_commit(oid)
                .context("Error al encontrar commit")?;

            let hash = oid.to_string();
            let message = commit.message().unwrap_or("").to_string();
            let author = commit.author().name().unwrap_or("").to_string();
            let timestamp = commit.time().seconds();

            commits.push(Commit {
                hash,
                message,
                author,
                timestamp,
            });
        }

        Ok(commits)
    }

    fn diff_file(&self, path: &str) -> Result<String> {
        let head = self.repo.head().ok();
        let head_tree = head
            .as_ref()
            .and_then(|h| h.peel_to_tree().ok());

        let diff = self.repo.diff_tree_to_workdir_with_index(
            head_tree.as_ref(),
            Some(
                git2::DiffOptions::new()
                    .pathspec(path)
            ),
        )?;

        let mut diff_text = String::new();
        diff.print(git2::DiffFormat::Patch, |_delta, _hunk, line| {
            let origin = line.origin();
            let content = std::str::from_utf8(line.content()).unwrap_or("");
            if !content.is_empty() && content.as_bytes()[0] == origin as u8 {
                diff_text.push(origin);
                diff_text.push_str(&content[1..]);
            } else {
                diff_text.push(origin);
                diff_text.push_str(content);
            }
            true
        })?;

        Ok(diff_text)
    }

    fn current_branch(&self) -> Result<Branch> {
        let head = self.repo.head()
            .context("Error al obtener HEAD")?;

        let name = head.shorthand()
            .unwrap_or("HEAD")
            .to_string();
        let upstream = self.repo.branch_upstream_name(&name)
            .ok()
            .and_then(|n| n.as_str().map(String::from));

        let (ahead, behind) = upstream.as_ref()
            .map(|_| {
                let local = self.repo.revparse_single(&name).ok()
                    .and_then(|obj| Some(obj.id()));
                let remote = self.repo.revparse_single(&format!("origin/{}", name)).ok()
                    .and_then(|obj| Some(obj.id()));

                match (local, remote) {
                    (Some(l), Some(r)) => {
                        let (a, b) = self.repo.graph_ahead_behind(l, r).unwrap_or((0, 0));
                        (a as usize, b as usize)
                    }
                    _ => (0, 0),
                }
            })
            .unwrap_or((0, 0));

        Ok(Branch {
            name,
            upstream,
            ahead,
            behind,
        })
    }
}

impl Git2Adapter {
    pub fn new(repo_path: &str) -> Result<Self> {
        let repo = Repository::open(repo_path)
            .context("No se pudo abrir el repositorio")?;
        Ok(Self { repo })
    }
}
