package main

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/namsral/flag"

	elogrus "github.com/dictor/echologrus"
	"github.com/go-git/go-git/v5"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// CustomValidator is struct validator for request input
type CustomValidator struct {
	validator *validator.Validate
}

var (
	gitTag, gitHash, buildDate string // build flags
	currentDoorStatus          bool   // latest door status
	latestChangeTime           int64  // latest door status changed timestamp
	// GlobalLogger is global default logger in whole program
	GlobalLogger elogrus.EchoLogger
)

// Validate is just renamed function of struct validate method
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

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
	e.Validator = &CustomValidator{validator: validator.New()}

	/* Get CLI flags */
	var (
		botToken, repoPath, storePath, listenPath string
		insertPeriod, uniquePeriod                int
	)
	flag.StringVar(&botToken, "token", "", "Bot's Token string")
	flag.StringVar(&repoPath, "repo", "https://github.com/ibarami/ibarami.github.io", "Repository path for check door status")
	flag.IntVar(&insertPeriod, "idelay", 60, "Getting web information task's period (second)")
	flag.IntVar(&uniquePeriod, "udelay", 86400, "Commit store uniqueness check task's period (second)")
	flag.StringVar(&storePath, "store", "./db.db", "Commit store file's path")
	flag.StringVar(&listenPath, "listen", ":80", "Listen address for web server")
	flag.Parse()

	/* Open DB */
	isNewStore, err := OpenStore(storePath)
	mustAction("store open", err)
	GlobalLogger.Infoln("commit store opened successfully!")

	/* Read ascii art */
	if err := readYsArt(); err != nil {
		GlobalLogger.WithError(err).Errorln("fail to read ascii art")
	}

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
		mustAction("append commits", AppendCommitToStore(commits...))
		mustAction("save store", SaveStore())
	}
	GlobalLogger.Infof("github repo cloned successfully!")
	go UpdateStatusLoop(repo, insertPeriod)
	go FixUniquenessLoop(uniquePeriod)

	/* Start web server */
	e.GET("/version", ReadVersion)
	e.GET("/commit", ReadCommit)
	e.GET("/latest", ReadLatestCommit)
	e.PUT("/log", UpdateLog)
	e.PUT("/heartbeat", UpdateHeartbeat)
	e.Logger.Fatal(e.Start(listenPath))
}

// UpdateStatusLoop is update door status in infinity loop
func UpdateStatusLoop(repo *git.Repository, delayPeriod int) {
	for {
		lcommit, err := SelectLatestCommit()
		if err != nil {
			GlobalLogger.Errorf("SelectLatestCommit: %s", err)
		}
		currentDoorStatus = lcommit.IsOpen

		scommit, err := SelectLatestStatus(lcommit.IsOpen)
		if err != nil {
			GlobalLogger.Errorf("SelectLatestCommit: %s", err)
		}
		latestChangeTime = scommit.CommitTime.Unix()

		commits, err := ListRepositoryCommits(repo, lcommit.CommitTime)
		if err != nil {
			GlobalLogger.Errorf("ListRepositoryCommits: %s", err)
		} else if len(commits) > 0 {
			if err := AppendCommitToStore(commits...); err != nil {
				GlobalLogger.Errorf("AppendStore: %s", err)
			}
			if err := SaveStore(); err != nil {
				GlobalLogger.Errorf("SaveStore: %s", err)
			}
		}

		time.Sleep(time.Second * time.Duration(delayPeriod))
	}
}

func FixUniquenessLoop(delayPeriod int) {
	for {
		beforeCnt := len(*GetAllCommits())
		err := FixUniqueness()
		afterCnt := len(*GetAllCommits())
		if err != nil {
			GlobalLogger.Errorf("FixUniqueness: %s", err)
		}
		GlobalLogger.Infof("FixUniqueness %d duplicated commits are fixed", beforeCnt-afterCnt)

		time.Sleep(time.Second * time.Duration(delayPeriod))
	}
}
