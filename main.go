package main

import (
	"database/sql"
	"flag"
	"fmt"
	"time"
)

var currentDoorStatus bool
var latestChangeTime int64

func main() {
	var bot_token string
	var insert_period int
	flag.StringVar(&bot_token, "t", "", "Bot's Token string")
	flag.IntVar(&insert_period, "d", 60, "Getting web information task's period (second)")
	flag.Parse()

	err := openDB("./db.db")
	if err != nil {
		fmt.Println("[DB Open Error]", err)
		return
	}
	fmt.Println("[DB opened successfully]")

	err = startBot(bot_token)
	if err != nil {
		fmt.Println("[Discord Bot Error]", err)
		return
	}
	fmt.Println("[Bot started successfully]")

	chan_change_status := make(chan bool)
	chan_change_time := make(chan int64)

	go func(chan_status chan bool, chan_time chan int64) {
		var last_status bool
		for {
			res_html, err := getRawHtml("https://ibarami.github.io/")
			if err != nil {
				fmt.Println("[HTTP Error]", err)
			} else {
				now_status := DoorStatus{time.Now().Unix(), isDoorOpen(res_html)}
				err = insertStatus(&now_status)
				if err != nil {
					fmt.Println("[DB Insert Error]", err)
				} else {
					fmt.Print("i")
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

			time.Sleep(time.Duration(insert_period) * time.Second)
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
