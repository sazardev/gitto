package gogit

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/ports"
)

type Adapter struct {
	repo    *git.Repository
	worktree *git.Worktree
}

func NewAdapter(repoPath string) (*Adapter, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	return &Adapter{
		repo:    repo,
		worktree: worktree,
	}, nil
}

func (a *Adapter) GetStatus() ([]entities.FileStatus, error) {
	status, err := a.worktree.Status()
	if err != nil {
		return nil, err
	}

	var files []entities.FileStatus
	for path := range status {
		fs := status.File(path)
		file := entities.NewFileStatus(path)

		if fs.Staging != git.Unmodified {
			file.IsStaged = true
			switch fs.Staging {
			case git.Modified:
				file.StagedStatus = entities.StatusModified
			case git.Added:
				file.StagedStatus = entities.StatusAdded
			case git.Deleted:
				file.StagedStatus = entities.StatusDeleted
			case git.Renamed:
				file.StagedStatus = entities.StatusRenamed
			case git.Copied:
				file.StagedStatus = entities.StatusCopied
			default:
				file.StagedStatus = entities.StatusUnmodified
			}
		}

		if status.IsUntracked(path) {
			file.UnstagedStatus = entities.StatusUntracked
			file.IsUntracked = true
		} else {
			switch fs.Worktree {
			case git.Modified:
				file.UnstagedStatus = entities.StatusModified
			case git.Added:
				file.UnstagedStatus = entities.StatusAdded
			case git.Deleted:
				file.UnstagedStatus = entities.StatusDeleted
			case git.Renamed:
				file.UnstagedStatus = entities.StatusRenamed
			default:
				file.UnstagedStatus = entities.StatusUnmodified
			}
		}

		files = append(files, file)
	}

	return files, nil
}

func (a *Adapter) Stage(paths []string) error {
	for _, path := range paths {
		_, err := a.worktree.Add(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Adapter) Unstage(paths []string) error {
	for range paths {
		err := a.worktree.Reset(&git.ResetOptions{
			Commit: plumbing.NewHash(""),
			Mode:   git.MixedReset,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Adapter) Commit(message string) error {
	_, err := a.worktree.Commit(message, &git.CommitOptions{
		All: true,
	})
	return err
}

func (a *Adapter) GetLog(limit int) ([]entities.Commit, error) {
	head, err := a.repo.Head()
	if err != nil {
		return nil, err
	}

	commits, err := a.repo.Log(&git.LogOptions{
		From: head.Hash(),
	})
	if err != nil {
		return nil, err
	}
	defer commits.Close()

	var result []entities.Commit
	for i := 0; i < limit; i++ {
		c, err := commits.Next()
		if err != nil {
			break
		}
		result = append(result, entities.NewCommit(
			c.Hash.String(),
			c.Message,
			c.Author.Name,
			c.Author.When,
		))
	}

	return result, nil
}

func (a *Adapter) GetDiff(filePath string) (string, error) {
	head, err := a.repo.Head()
	if err != nil {
		return "", err
	}

	commit, err := a.repo.CommitObject(head.Hash())
	if err != nil {
		return "", err
	}

	tree, err := commit.Tree()
	if err != nil {
		return "", err
	}

	file, err := tree.File(filePath)
	if err != nil {
		return "", err
	}

	return file.Contents()
}

func (a *Adapter) Push() error {
	return a.repo.Push(&git.PushOptions{})
}

func (a *Adapter) Pull() error {
	return a.worktree.Pull(&git.PullOptions{})
}

func (a *Adapter) GetCurrentBranch() (string, error) {
	head, err := a.repo.Head()
	if err != nil {
		return "", err
	}

	return head.Name().Short(), nil
}

var _ ports.GitProvider = (*Adapter)(nil)