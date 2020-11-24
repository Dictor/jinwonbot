package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"log"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

type (
	Commit struct {
		CommitTime time.Time // Pushed time
		IsOpen     bool
		EventTime  time.Time // Time in commit message
	}
)

func parseCommitMessage(message string) (eventTime time.Time, isOpen bool, err error) {
	words := strings.Split(message, " ")
	if len(words) != 3 {
		err = errors.New(fmt.Sprintf("parseCommitMessage: expected 3 words in commit message, but %d existed. (msg: %s)", len(words), message))
		return
	}

	if words[2] == "닫힘" {
		isOpen = false
	} else if words[2] == "열림" {
		isOpen = true
	} else {
		err = errors.New(fmt.Sprintf("parseCommitMessage: unknown door status message : %s. (msg: %s)", words[2], message))
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

func CloneGitRepository(url string) (*git.Repository, error) {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:        url,
		RemoteName: "master",
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func ListRepositoryCommits(repo *git.Repository) ([]*Commit, error) {
	headRef, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commitIter, err := repo.Log(&git.LogOptions{From: headRef.Hash()})
	if err != nil {
		return nil, err
	}

	res := []*Commit{}
	err = commitIter.ForEach(func(c *object.Commit) error {
		etime, open, err := parseCommitMessage(c.Message)
		if err != nil {
			log.Println(err)
			return nil
		}
		res = append(res, &Commit{
			CommitTime: c.Committer.When,
			IsOpen:     open,
			EventTime:  etime,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, err
}
