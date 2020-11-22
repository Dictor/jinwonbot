package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFullCommits(t *testing.T) {
	commits, err := getFullCommits()
	assert.NoError(t, err)
	fmt.Println(commits)
}
