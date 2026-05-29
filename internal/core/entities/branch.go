package entities

type Branch struct {
	Name   string
	IsHead bool
	IsRemote bool
}

func NewBranch(name string, isHead, isRemote bool) Branch {
	return Branch{
		Name:   name,
		IsHead: isHead,
		IsRemote: isRemote,
	}
}