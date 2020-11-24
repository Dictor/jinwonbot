package main

import (
	"flag"
	"log"
	"time"

	"github.com/dictor/justlog"
)

const version string = "v1.0.2"

var currentDoorStatus bool
var latestChangeTime int64

func checkError(explain string, err error) {
	if err != nil {
		log.Panicf("[%s → Fatal error!] %v", explain, err)
	}
}

func main() {
	/* Set logging */
	log_path := justlog.MustPath(justlog.SetPath())
	defer (justlog.MustStream(justlog.SetStream(log_path))).Close()

	/* Get CLI flags */
	var bot_token, web_path string
	var insert_period int
	flag.StringVar(&bot_token, "t", "", "Bot's Token string")
	flag.StringVar(&web_path, "w", "https://ibarami.github.io/", "Web path for check door status")
	flag.IntVar(&insert_period, "d", 60, "Getting web information task's period (second)")
	flag.Parse()

	/* Open DB */
	isNewStore, err := OpenStore("./db.db")
	checkError("DB Open", err)
	log.Println("[DB opened successfully]")

	/* Start discord bot */
	err = startBot(bot_token)
	checkError("Starting bot", err)
	log.Println("[Bot started successfully]")

	/* Get github information */
	repo, err := CloneGitRepository("https://github.com/ibarami/ibarami.github.io")
	checkError("Clone repo", err)
	if isNewStore {
		commits, err := ListRepositoryCommits(repo, time.Time{})
		checkError("Get commits", err)
		checkError("Update DB", AppendStore(commits...))
		checkError("Save DB", SaveStore())
	}

	chan_change_status := make(chan bool)
	chan_change_time := make(chan int64)

	go func(chan_status chan bool, chan_time chan int64) {
		var last_status bool
		for {

			commit, err := SelectLatestCommit()
			if err != nil {
				log.Println("[Commit retrieve error]", err)
			}
			repo, err = CloneGitRepository("https://github.com/ibarami/ibarami.github.io")
			if err != nil {
				log.Println("[repo clone error]", err)
			}
			commits, err := ListRepositoryCommits(repo, commit.CommitTime)
			if err != nil {
				log.Println("[Commit select error]", err)
			}
			if len(commits) > 0 {
				err := AppendStore(commits...)
				if err != nil {
					log.Println("[append store error]", err)
				}
				err = SaveStore()
				if err != nil {
					log.Println("[save store error]", err)
				}
			}

			if err != nil {
				log.Println("[DB Select Error]", err)
			}
			if commit.IsOpen != last_status {
				last_status = commit.IsOpen
				chan_status <- commit.IsOpen
			}

			scommit, err := SelectLatestStatus(last_status)
			if err != nil {
				log.Println("[DB Select Error]", err)
			}
			chan_time <- scommit.CommitTime.Unix()

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
