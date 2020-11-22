package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type (
	Commit struct {
		CommitTime time.Time // Pushed time
		IsOpen     bool
		EventTime  time.Time // Time in commit message
	}
)

const apiCallDelay = 100 * time.Millisecond

func getFullCommits() ([]*Commit, error) {
	page := 0
	res := []*Commit{}
	for {
		data, err := callCommitAPI(page)
		if err != nil {
			return nil, err
		}
		if len(data) == 0 {
			break
		}

		for _, commit := range data {
			commitData := commit.(map[string]interface{})["commit"].(map[string]interface{})
			commitMessage := commitData["message"].(string)
			etime, open, err := parseCommitMessage(commitMessage)
			if err != nil {
				return nil, err
			}
			ctime, err := time.Parse(time.RFC3339, commitData["committer"].(map[string]interface{})["date"].(string))
			if err != nil {
				return nil, err
			}
			res = append(res, &Commit{CommitTime: ctime, IsOpen: open, EventTime: etime})
		}
	}
	return res, nil
}

func parseCommitMessage(message string) (eventTime time.Time, isOpen bool, err error) {
	words := strings.Split(message, " ")
	if len(words) != 3 {
		err = errors.New(fmt.Sprintf("parseCommitMessage: expected 3 words in commit message, but %d existed.", len(words)))
		return
	}

	if words[2] == "닫힘" {
		isOpen = false
	} else if words[2] == "열림" {
		isOpen = true
	} else {
		err = errors.New(fmt.Sprintf("parseCommitMessage: unknown door status message : %s", words[2]))
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

func callCommitAPI(page int) ([]interface{}, error) {
	client := resty.New()
	resp, err := client.R().
		EnableTrace().
		Get("https://api.github.com/repos/ibarami/ibarami.github.io/commits?page=" + strconv.Itoa(page))
	if err != nil {
		return nil, err
	}
	body := []interface{}{}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, err
	}
	return body, nil
}
