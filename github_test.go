package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFullCommits(t *testing.T) {
	repo, err := CloneGitRepository("https://github.com/ibarami/ibarami.github.io")
	assert.NoError(t, err)
	commits, err := ListRepositoryCommits(repo)
	assert.NoError(t, err)
	fmt.Printf("%d commits retireved.", len(commits))
}
