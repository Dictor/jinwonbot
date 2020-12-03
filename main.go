package main

import (
	"log"
	"time"

	"github.com/namsral/flag"

	elogrus "github.com/dictor/echologrus"
	"github.com/go-git/go-git/v5"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var (
	gitTag, gitHash, buildDate string // build flags
	currentDoorStatus          bool   // latest door status
	latestChangeTime           int64  // latest door status changed timestamp
	// GlobalLogger is global default logger in whole program
	GlobalLogger elogrus.EchoLogger
)

func mustAction(action string, err error) {
	if err != nil {
		GlobalLogger.WithFields(logrus.Fields{
			"action": action,
			"err":    err,
		}).Panic("necessary action failed")
	}
}

func main() {
	/* Init webserver and logger */
	e := echo.New()
	GlobalLogger = elogrus.Attach(e)
	GlobalLogger.Infof("jinwonbot %s (%s) - %s", gitTag, gitHash, buildDate)

	/* Get CLI flags */
	var (
		botToken, repoPath, storePath, listenPath string
		insertPeriod                              int
	)
	flag.StringVar(&botToken, "token", "", "Bot's Token string")
	flag.StringVar(&repoPath, "repo", "https://github.com/ibarami/ibarami.github.io", "Repository path for check door status")
	flag.IntVar(&insertPeriod, "delay", 60, "Getting web information task's period (second)")
	flag.StringVar(&storePath, "store", "./db.db", "Commit store file's path")
	flag.StringVar(&listenPath, "listen", ":80", "Listen address for web server")
	flag.Parse()

	/* Open DB */
	isNewStore, err := OpenStore(storePath)
	mustAction("store open", err)
	GlobalLogger.Infoln("commit store opened successfully!")

	/* Start discord bot */
	err = startBot(botToken)
	mustAction("start bot", err)
	GlobalLogger.Infoln("discord bot started successfully!")

	/* Get github information */
	repo, err := CloneGitRepository(repoPath)
	mustAction("clone repo", err)
	if isNewStore {
		GlobalLogger.Infoln("there isn't saved commit store. creating...")
		commits, err := ListRepositoryCommits(repo, time.Time{})
		mustAction("get commits", err)
		GlobalLogger.Infof("%d commits retieved!", len(commits))
		mustAction("append commits", AppendStore(commits...))
		mustAction("save store", SaveStore())
	}
	GlobalLogger.Infof("github repo cloned successfully!")
	go UpdateStatusLoop(repo, insertPeriod)

	/* Start web server */
	e.GET("/version", ReadVersion)
	e.GET("/commit", ReadCommit)
	e.GET("/latest", ReadLatestCommit)
	e.Logger.Fatal(e.Start(listenPath))
}

// UpdateStatusLoop is update door status in infinity loop
func UpdateStatusLoop(repo *git.Repository, delayPeriod int) {
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

		time.Sleep(time.Second * time.Duration(delayPeriod))
	}
}
