package entities

import "time"

type Commit struct {
	Hash       string
	ShortHash  string
	Message    string
	Author     string
	AuthorDate time.Time
}

func NewCommit(hash, message, author string, authorDate time.Time) Commit {
	return Commit{
		Hash:       hash,
		ShortHash:  hash[:7],
		Message:    message,
		Author:     author,
		AuthorDate: authorDate,
	}
}