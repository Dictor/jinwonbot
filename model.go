package main

import (
	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters/sqlite"
	"os"
)

type DoorStatus struct {
	Time   int64 `db:"timestamp,key"`
	Status bool  `db:"status"`
}

var currentDB *godb.DB

func openDB(path string) error {
	var err error
	var table_init bool = false
	if fileExist(path) {
		table_init = true
	}

	currentDB, err = godb.Open(sqlite.Adapter, path)
	if !table_init {
		_, interr := currentDB.CurrentDB().Exec("CREATE TABLE DoorStatus(timestamp INTEGER PRIMARY KEY NOT NULL, status TEXT NOT NULL)")
		if interr != nil {
			return interr
		}
	}

	return err
}

func closeDB() error {
	err := currentDB.Close()
	return err
}

func insertStatus(status *DoorStatus) error {
	return currentDB.Insert(status).Do()
}

func getLatestStatus() (*DoorStatus, error) {
	now_status := DoorStatus{}
	err := currentDB.Select(&now_status).OrderBy("timestamp DESC").Limit(1).Do()
	return &now_status, err
}

func getLatestConditionStatus(condition bool) (*DoorStatus, error) {
	now_status := DoorStatus{}
	err := currentDB.Select(&now_status).Where("status = ?", condition).OrderBy("timestamp DESC").Limit(1).Do()
	return &now_status, err
}

func fileExist(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
