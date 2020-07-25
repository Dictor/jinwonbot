package main

import (
	"github.com/timshannon/badgerhold"
	"os"
	"time"
	"fmt"
)

type DoorStatus struct {
	Time   int64
	Status bool
}

var currentStore *badgerhold.Store

func openDB(path string) error {
	var err error
	opt := badgerhold.DefaultOptions
	opt.Dir = path
	opt.ValueDir = path
	currentStore, err = badgerhold.Open(opt)
	return err
}

func closeDB() error {
	err := currentStore.Close()
	return err
}

func insertStatus(status *DoorStatus) error {
	return currentStore.Insert(fmt.Sprintf("status,%d", time.Now().Unix()), status)
}

func getLatestStatus() (*DoorStatus, error) {
	nowStatus := DoorStatus{}
	res, err := currentStore.FindAggregate(&nowStatus, nil, "Time")
	if len(res) < 1 {
		return nil, err
	}
	res[0].Max("Time", &nowStatus)
	return &nowStatus, err
}

func getLatestConditionStatus(condition bool) (*DoorStatus, error) {
	nowStatus := []DoorStatus{}
	err := currentStore.Find(&nowStatus, badgerhold.Where("Status").Eq(condition).SortBy("Time").Limit(1))
	if len(nowStatus) < 1 {
		return nil, err
	}
	return &nowStatus[0], err
}

func fileExist(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
