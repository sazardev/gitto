package ports

import "github.com/sazardev/gitto/internal/core/entities"

type GitProvider interface {
	GetStatus() ([]entities.FileStatus, error)
	Stage(paths []string) error
	Unstage(paths []string) error
	Commit(message string) error
	GetLog(limit int) ([]entities.Commit, error)
	GetDiff(path string) (string, error)
	Push() error
	Pull() error
	GetCurrentBranch() (string, error)
}