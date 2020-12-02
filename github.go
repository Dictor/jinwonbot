package main

import (
	"fmt"
	"strings"
	"time"

	"log"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

type (
	// Commit is model of each git commits.
	Commit struct {
		CommitTime time.Time // Pushed time
		IsOpen     bool
		EventTime  time.Time // Time in commit message
		Hash       string
	}
)

func (c Commit) String() string {
	return fmt.Sprintf("감지 시간 %s → 커밋 시간 %s 개방?=%t (해시 %s)", c.EventTime, c.CommitTime, c.IsOpen, c.Hash[0:6])
}

func parseCommitMessage(message string) (eventTime time.Time, isOpen bool, err error) {
	words := strings.Split(message, " ")
	if len(words) != 3 {
		err = fmt.Errorf("parseCommitMessage: expected 3 words in commit message, but %d existed. (msg: %s)", len(words), message)
		return
	}

	if words[2] == "닫힘" {
		isOpen = false
	} else if words[2] == "열림" {
		isOpen = true
	} else {
		err = fmt.Errorf("parseCommitMessage: unknown door status message : %s. (msg: %s)", words[2], message)
		return
	}

	etime, terr := time.Parse("2006-01-02 15:04:05", words[0]+" "+words[1])
	if terr != nil {
		err = terr
		return
	}
	eventTime = etime
	return
}

// CloneGitRepository returns repository object from given git repository url
func CloneGitRepository(url string) (*git.Repository, error) {
	repo, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:        url,
		RemoteName: "master",
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// ListRepositoryCommits returns whole commits in given repository
func ListRepositoryCommits(repo *git.Repository, since time.Time) ([]*Commit, error) {
	w, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	if err := w.Pull(&git.PullOptions{RemoteName: "master"}); err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, err
	}

	headRef, err := repo.Head()
	if err != nil {
		return nil, err
	}
	commitIter, err := repo.Log(&git.LogOptions{From: headRef.Hash(), Since: &since})
	if err != nil {
		return nil, err
	}

	res := []*Commit{}
	if err := commitIter.ForEach(func(c *object.Commit) error {
		etime, open, err := parseCommitMessage(c.Message)
		if err != nil {
			log.Println(err)
			return nil
		}
		res = append(res, &Commit{
			CommitTime: c.Committer.When,
			IsOpen:     open,
			EventTime:  etime,
			Hash:       c.Hash.String(),
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return res, nil
}
