package entities

type DiffHunk struct {
	Header     string
	OldStart   int
	OldLines   int
	NewStart   int
	NewLines   int
	Lines      []DiffLine
	Raw        []string
}

type DiffLine struct {
	Content  string
	Type     DiffLineType
	OldLineNo int
	NewLineNo int
}

type DiffLineType int

const (
	DiffLineContext DiffLineType = iota
	DiffLineAdded
	DiffLineDeleted
	DiffLineHeader
)

type Diff struct {
	FilePath string
	Hunks    []DiffHunk
}

func NewDiff(filePath string) Diff {
	return Diff{
		FilePath: filePath,
		Hunks:    []DiffHunk{},
	}
}