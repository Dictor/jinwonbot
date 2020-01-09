package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"
)

var currentDoorStatus bool
var latestChangeTime int64

func checkError(explain string, err error) {
	if err != nil {
		log.Panicf("[%s → Fatal error!] %v", explain, err)
	}
}

func main() {
	/* Set logging */
	log_path, _ := setLogPath() // Ignore error in this line is finally cause panic in set log stream anyway.
	fp, err := setLogStream(log_path)
	defer fp.Close()
	checkError("Set logging", err)

	/* Get CLI flags */
	var bot_token, web_path string
	var insert_period int
	flag.StringVar(&bot_token, "t", "", "Bot's Token string")
	flag.StringVar(&web_path, "w", "https://ibarami.github.io/", "Web path for check door status")
	flag.IntVar(&insert_period, "d", 60, "Getting web information task's period (second)")
	flag.Parse()

	/* Open DB */
	err = openDB("./db.db")
	checkError("DB Open", err)
	fmt.Println("[DB opened successfully]")

	/* Start discord bot */
	err = startBot(bot_token)
	checkError("Starting bot", err)
	fmt.Println("[Bot started successfully]")

	chan_change_status := make(chan bool)
	chan_change_time := make(chan int64)

	go func(chan_status chan bool, chan_time chan int64) {
		var last_status bool
		for {
			res_html, err := getRawHtml(web_path)
			if err != nil {
				log.Println("[HTTP Error]", err)
			} else {
				now_status := DoorStatus{time.Now().Unix(), isDoorOpen(res_html)}
				err = insertStatus(&now_status)
				if err != nil {
					log.Println("[DB Insert Error]", err)
				} else {
					log.Print("i")
				}
			}

			res_last, err := getLatestStatus()
			if err != nil {
				if err != sql.ErrNoRows {
					log.Println("[DB Select Error]", err)
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
					log.Println("[DB Select Error]", err)
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
			log.Println("[Status changed] →", currentDoorStatus)
		case latestChangeTime = <-chan_change_time:
		}
	}
}
