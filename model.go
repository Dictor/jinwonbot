package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"
)

type (
	// CommitStore is commits storing json store
	CommitStore struct {
		Version int64
		Commits []*Commit
		Lock    *sync.Mutex `json:"-"` //Write sync mutex
	}
)

var (
	currentStore      *CommitStore
	currentStorePath  string
	lastUpdateVersion int64
)

// OpenStore open commit store. When there is no file in path, create new store on the path.
func OpenStore(path string) (isNew bool, openError error) {
	if currentStore != nil {
		openError = errors.New("commit store is already opened")
		return
	}

	if _, err := os.Stat(path); err == nil {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			openError = err
			return
		}
		if err := json.Unmarshal(data, &currentStore); err != nil {
			openError = err
			return
		}
		currentStorePath = path
		isNew = false
	} else if os.IsNotExist(err) {
		currentStore = &CommitStore{}
		data, err := json.Marshal(currentStore)
		if err != nil {
			openError = err
			return
		}
		if err := ioutil.WriteFile(path, data, os.ModePerm); err != nil {
			openError = err
			return
		}
		currentStorePath = path
		isNew = true
	} else {
		openError = err
		return
	}
	currentStore.Lock = &sync.Mutex{}
	return
}

// AppendStore append given commits in commit store
func AppendStore(commits ...*Commit) error {
	if currentStore == nil {
		return errors.New("there is no opened store")
	}

	currentStore.Lock.Lock()
	currentStore.Commits = append(currentStore.Commits, commits...)
	currentStore.Version++
	currentStore.Lock.Unlock()
	return nil
}

// SaveStore save commit store to file system
func SaveStore() error {
	if currentStore == nil {
		return errors.New("there is no opened store")
	}
	if currentStore.Version <= lastUpdateVersion {
		return nil
	}

	f, err := os.OpenFile(currentStorePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	err = f.Truncate(0)
	if err != nil {
		return err
	}

	data, err := json.Marshal(currentStore)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

// CloseStore close and free commit store
func CloseStore() error {
	if err := SaveStore(); err != nil {
		return err
	}
	currentStore = nil
	lastUpdateVersion = 0
	return nil
}

// SelectLatestCommit returns latest commit
func SelectLatestCommit() (*Commit, error) {
	tcommit, err := SelectLatestStatus(true)
	if err != nil {
		return nil, err
	}
	fcommit, err := SelectLatestStatus(false)
	if err != nil {
		return nil, err
	}

	if tcommit.CommitTime.After(fcommit.CommitTime) {
		return tcommit, nil
	} else {
		return fcommit, nil
	}
}

// SelectLatestStatus returns latest commits matched with given condition
func SelectLatestStatus(status bool) (*Commit, error) {
	if len(currentStore.Commits) < 2 {
		return nil, errors.New("there is no commit to select in store")
	}

	var (
		latestCommit *Commit = currentStore.Commits[0]
	)
	for _, c := range currentStore.Commits {
		if c.IsOpen == status {
			if c.CommitTime.After(latestCommit.CommitTime) {
				latestCommit = c
			}
		}
	}

	return latestCommit, nil
}

// GetAllCommits return slice of whole commits in store
func GetAllCommits() *[]*Commit {
	return &currentStore.Commits
}

func FixUniqueness() error {
	if currentStore == nil {
		return errors.New("there is no opened store")
	}

	existHash := []string{}
	uniqueCommits := []*Commit{}
	cs := GetAllCommits()
	for _, c := range *cs {
		if !contains(existHash, c.Hash) {
			existHash = append(existHash, c.Hash)
			uniqueCommits = append(uniqueCommits, c)
		}
	}
	currentStore.Lock.Lock()
	currentStore.Commits = uniqueCommits
	currentStore.Lock.Unlock()
	return nil
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
