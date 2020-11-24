package main

import (
	"flag"
	"log"
	"time"

	"github.com/dictor/justlog"
)

const version string = "v1.0.2"

var (
	currentDoorStatus bool
	latestChangeTime  int64
)

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
	var (
		botToken, repoPath string
		insertPeriod       int
	)
	flag.StringVar(&botToken, "t", "", "Bot's Token string")
	flag.StringVar(&repoPath, "r", "https://github.com/ibarami/ibarami.github.io", "Repository path for check door status")
	flag.IntVar(&insertPeriod, "d", 60, "Getting web information task's period (second)")
	flag.Parse()

	/* Open DB */
	isNewStore, err := OpenStore("./db.db")
	checkError("store open", err)
	log.Println("commit store opened successfully!")

	/* Start discord bot */
	err = startBot(botToken)
	checkError("starting bot", err)
	log.Println("discord bot started successfully!")

	/* Get github information */
	repo, err := CloneGitRepository(repoPath)
	checkError("clone repo", err)
	if isNewStore {
		log.Println("there isn't saved commit store. creating...")
		commits, err := ListRepositoryCommits(repo, time.Time{})
		checkError("get commits", err)
		log.Printf("%d commits retieved!", len(commits))
		checkError("append store", AppendStore(commits...))
		checkError("save store", SaveStore())
	}
	log.Println("github repo cloned successfully!")

	for {
		lcommit, err := SelectLatestCommit()
		if err != nil {
			log.Printf("SelectLatestCommit: %s", err)
		}
		currentDoorStatus = lcommit.IsOpen

		scommit, err := SelectLatestStatus(lcommit.IsOpen)
		if err != nil {
			log.Printf("SelectLatestCommit: %s", err)
		}
		latestChangeTime = scommit.CommitTime.Unix()

		commits, err := ListRepositoryCommits(repo, lcommit.CommitTime)
		if err != nil {
			log.Printf("ListRepositoryCommits: %s", err)
		} else if len(commits) > 0 {
			if err := AppendStore(commits...); err != nil {
				log.Printf("AppendStore: %s", err)
			}
			if err := SaveStore(); err != nil {
				log.Printf("SaveStore: %s", err)
			}
		}

		time.Sleep(time.Second * time.Duration(insertPeriod))
	}
}
