package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

type (
	// CommitStore is commits storing json store
	CommitStore struct {
		Version   int64
		Commits   []*Commit
		Hearbeats *map[string]string
		Logs      *map[string]string
		Info      *map[string]string
		Lock      *sync.Mutex `json:"-"` //Write sync mutex
	}

	LogUpdateRequest struct {
		Level string `json:"level" validate:"required"`
		Data  string `json:"data" validate:"required"`
	}

	InfoKey string
)

var (
	currentStore            *CommitStore
	currentStorePath        string
	lastUpdateVersion       int64
	currentStoreDebugString string
)

const (
	CallCount InfoKey = InfoKey("call_count")
)

// OpenStore open commit store. When there is no file in path, create new store on the path.
func OpenStore(path string) (isNew bool, openError error) {
	if currentStore != nil {
		openError = errors.New("commit store is already opened")
		return
	}

	before := time.Now()
	storeSize := 0
	if _, err := os.Stat(path); err == nil {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			openError = err
			return
		}
		storeSize = len(data)
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

	if currentStore.Hearbeats == nil {
		currentStore.Hearbeats = &map[string]string{}
	}

	if currentStore.Logs == nil {
		currentStore.Logs = &map[string]string{}
	}

	if currentStore.Info == nil {
		currentStore.Info = &map[string]string{}
	}

	currentStore.Lock = &sync.Mutex{}
	after := time.Since(before)
	currentStoreDebugString = fmt.Sprintf("크기는 %d 바이트, 여는데 %d ms 걸렸습니다.", storeSize, after.Milliseconds())
	return
}

// AppendCommitToStore append given commits in commit store
func AppendCommitToStore(commits ...*Commit) error {
	if currentStore == nil {
		return errors.New("there is no opened store")
	}

	currentStore.Lock.Lock()
	currentStore.Commits = append(currentStore.Commits, commits...)
	currentStore.Version++
	currentStore.Lock.Unlock()
	return nil
}

func AppendLogToStore(ip string, level string, data string) error {
	if currentStore == nil {
		return errors.New("there is no opened store")
	}
	currentStore.Lock.Lock()

	logs := (*currentStore.Logs)
	log, exist := logs[ip]
	line := fmt.Sprintf("[%s] %s %s", time.Now().Format("06/01/02 15:04:05"), level, data)
	if !exist {
		log = line
	} else {
		log = strings.Join([]string{log, line}, "\n")
	}
	logs[ip] = log
	currentStore.Logs = &logs

	currentStore.Version++
	currentStore.Lock.Unlock()
	return nil
}

func SetInfoToStore(key InfoKey, data string) error {
	if currentStore == nil {
		return errors.New("there is no opened store")
	}
	currentStore.Lock.Lock()

	info := (*currentStore.Info)
	info[string(key)] = data
	currentStore.Info = &info

	currentStore.Version++
	currentStore.Lock.Unlock()
	return nil
}

func UpdateHeartbeatToStore(ip string) error {
	if currentStore == nil {
		return errors.New("there is no opened store")
	}
	currentStore.Lock.Lock()

	beatTime := (*currentStore.Hearbeats)
	beatTime[ip] = time.Now().Format(time.RFC3339)
	currentStore.Hearbeats = &beatTime
	currentStore.Version++

	currentStore.Lock.Unlock()
	return nil
}

func GetStoreVersion() int64 {
	return currentStore.Version
}

func GetStoreDebugString() string {
	return currentStoreDebugString
}

func GetInfo(key InfoKey) string {
	if v, ok := (*currentStore.Info)[string(key)]; ok {
		return v
	} else {
		return ""
	}
}

// return key = ip address, value = log string
func GetLogString() map[string]string {
	logs := *currentStore.Logs
	constraintLogs := map[string]string{}

	for k, v := range logs {
		constraintLogs[k] = v
	}
	return constraintLogs
}

func GetLatestHeartbeat() *struct {
	string
	int64
} {
	list := make(map[string]int64, 8)
	hbs := *currentStore.Hearbeats
	for ip, t := range hbs {
		pt, err := time.Parse(time.RFC3339, t)
		if err == nil {
			list[ip] = int64(time.Since(pt).Seconds())
		}
	}

	if len(list) == 0 {
		return nil
	}

	mink := ""
	minv := int64(math.MaxInt64)
	for ip, t := range list {
		if t <= minv {
			mink = ip
			minv = t
		}
	}
	return &struct {
		string
		int64
	}{mink, minv}
}

func GetHeartbeatString() string {
	hbs := *currentStore.Hearbeats
	str := ""

	for ip, t := range hbs {
		pt, err := time.Parse(time.RFC3339, t)
		if err == nil {
			t = formatSecond(int64(time.Since(pt).Seconds()))
		}
		str += fmt.Sprintf("[%s] : %s\n", ip, t)
	}

	if str == "" {
		str = "내용 없음"
	}

	return str
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
