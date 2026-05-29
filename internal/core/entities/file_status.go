package entities

type FileStatusType int

const (
	StatusUnmodified FileStatusType = iota
	StatusModified
	StatusAdded
	StatusDeleted
	StatusRenamed
	StatusCopied
	StatusUntracked
	StatusIgnored
	StatusStaged
	StatusUnstaged
)

func (s FileStatusType) String() string {
	switch s {
	case StatusUnmodified:
		return "unmodified"
	case StatusModified:
		return "modified"
	case StatusAdded:
		return "added"
	case StatusDeleted:
		return "deleted"
	case StatusRenamed:
		return "renamed"
	case StatusCopied:
		return "copied"
	case StatusUntracked:
		return "untracked"
	case StatusIgnored:
		return "ignored"
	case StatusStaged:
		return "staged"
	case StatusUnstaged:
		return "unstaged"
	default:
		return "unknown"
	}
}

type FileStatus struct {
	Path          string
	StagedStatus  FileStatusType
	UnstagedStatus FileStatusType
	IsStaged      bool
	IsUntracked   bool
	IsIgnored     bool
}

func NewFileStatus(path string) FileStatus {
	return FileStatus{
		Path: path,
	}
}