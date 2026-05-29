package ui

import "github.com/sazardev/gitto/internal/core/entities"

type CommitSuccess struct{}

type CommitError struct {
	Err error
}

type CloseCommitView struct{}

type RefreshStatus struct{}

type PushSuccess struct{}

type PushError struct {
	Err error
}

type PullSuccess struct{}

type PullError struct {
	Err error
}

type LogLoaded struct {
	Commits []entities.Commit
}

type LogError struct {
	Err error
}

type DashboardLoaded struct {
	Dashboard entities.Dashboard
}

type DashboardError struct {
	Err error
}
