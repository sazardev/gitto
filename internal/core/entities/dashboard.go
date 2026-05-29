package entities

type Dashboard struct {
	CurrentBranch string
	Branches      []Branch
	Commits       []Commit
	Files         []FileStatus
	StagedCount   int
	UnstagedCount int
	UntrackedCount int
}

func NewDashboard() Dashboard {
	return Dashboard{
		Branches: []Branch{},
		Commits:  []Commit{},
		Files:    []FileStatus{},
	}
}
