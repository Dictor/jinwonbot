package main

import (
	"database/sql"
	"fmt"
	"time"
)

var currentDoorStatus bool
var latestChangeTime int64

func main() {
	err := openDB("./db.db")
	if err != nil {
		fmt.Println("[DB Open Error]", err)
		return
	}

	chan_change_status := make(chan bool)
	chan_change_time := make(chan int64)

	go func(chan_status chan bool, chan_time chan int64) {
		var last_status bool
		for {
			res_html, err := getRawHtml("https://ibarami.github.io/")
			if err != nil {
				fmt.Println(err)
			} else {
				now_status := DoorStatus{time.Now().Unix(), isDoorOpen(res_html)}
				err = insertStatus(&now_status)
				if err != nil {
					fmt.Println("[DB Insert Error]", err)
				}
			}

			res_last, err := getLatestStatus()
			if err != nil {
				if err != sql.ErrNoRows {
					fmt.Println("[DB Select Error]", err)
				}
			} else {
				if res_last.Status != last_status {
					last_status = res_last.Status
					chan_status <- res_last.Status
				}
			}

			res_time, err := getLatestConditionStatus(!last_status)
			if err != nil {
				if err != sql.ErrNoRows {
					fmt.Println("[DB Select Error]", err)
				}
			} else {
				chan_time <- res_time.Time
			}

			time.Sleep(10 * time.Second)
		}
	}(chan_change_status, chan_change_time)

	for {
		select {
		case currentDoorStatus = <-chan_change_status:
			fmt.Println("[Status changed] →", currentDoorStatus)
		case latestChangeTime = <-chan_change_time:
		}
	}
}
