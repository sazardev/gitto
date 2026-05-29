package gogit

import (
	"os/exec"
	"strings"

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
	return a.worktree.Reset(&git.ResetOptions{
		Commit: plumbing.NewHash(""),
		Mode:   git.MixedReset,
		Files:  paths,
	})
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

func (a *Adapter) GetDiff(filePath string, staged bool) (*entities.Diff, error) {
	result := entities.NewDiff(filePath)

	var cmd *exec.Cmd
	if staged {
		cmd = exec.Command("git", "diff", "--cached", "--", filePath)
	} else {
		cmd = exec.Command("git", "diff", "--", filePath)
	}
	cmd.Dir = a.worktree.Filesystem.Root()

	output, err := cmd.Output()
	if err != nil {
		return &result, nil
	}

	diffStr := string(output)
	lines := parseDiffOutput(diffStr)
	if len(lines) == 0 {
		return &result, nil
	}

	result.Hunks = []entities.DiffHunk{{Lines: lines}}
	return &result, nil
}

func parseDiffOutput(diffStr string) []entities.DiffLine {
	var lines []entities.DiffLine
	diffLines := strings.Split(diffStr, "\n")

	for _, line := range diffLines {
		if len(line) == 0 {
			continue
		}

		var diffLine entities.DiffLine
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			diffLine = entities.DiffLine{
				Content: line,
				Type:    entities.DiffLineHeader,
			}
		case strings.HasPrefix(line, "+"):
			diffLine = entities.DiffLine{
				Content: line[1:],
				Type:    entities.DiffLineAdded,
			}
		case strings.HasPrefix(line, "-"):
			diffLine = entities.DiffLine{
				Content: line[1:],
				Type:    entities.DiffLineDeleted,
			}
		case strings.HasPrefix(line, "@"):
			diffLine = entities.DiffLine{
				Content: line,
				Type:    entities.DiffLineHeader,
			}
		default:
			diffLine = entities.DiffLine{
				Content: line,
				Type:    entities.DiffLineContext,
			}
		}
		lines = append(lines, diffLine)
	}

	return lines
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

func (a *Adapter) GetBranches() ([]entities.Branch, error) {
	head, err := a.repo.Head()
	if err != nil {
		return nil, err
	}

	var branches []entities.Branch

	iter, err := a.repo.Branches()
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	err = iter.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		isHead := ref.Hash() == head.Hash()
		branches = append(branches, entities.NewBranch(name, isHead, false))
		return nil
	})
	if err != nil {
		return nil, err
	}

	remoteIter, err := a.repo.References()
	if err != nil {
		return branches, nil
	}
	defer remoteIter.Close()

	err = remoteIter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsRemote() {
			name := ref.Name().Short()
			branches = append(branches, entities.NewBranch(name, false, true))
		}
		return nil
	})

	return branches, nil
}

var _ ports.GitProvider = (*Adapter)(nil)