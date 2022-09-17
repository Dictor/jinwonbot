package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type errorReason string

const (
	internalServerError errorReason = "internal_server_error"
	invalidParam        errorReason = "invalid_param"
)

var errorReasonToCode map[errorReason]int = map[errorReason]int{
	internalServerError: http.StatusInternalServerError,
	invalidParam:        http.StatusBadRequest,
}

func ReadVersion(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"name":    "jinwonbot",
		"version": fmt.Sprintf("%s (%s) - %s", gitTag, gitHash, buildDate),
	})
}

func ReadLatestCommit(c echo.Context) error {
	statusCondition := c.QueryParam("status")
	var (
		status *Commit
		err    error
	)
	switch statusCondition {
	case "open":
		status, err = SelectLatestStatus(true)
	case "close":
		status, err = SelectLatestStatus(false)
	default:
		status, err = SelectLatestCommit()
	}
	if err != nil {
		return errorResponse(c, internalServerError)
	}
	return c.JSON(http.StatusOK, status)
}

func ReadCommit(c echo.Context) error {
	commits := GetAllCommits()
	commitslen := len(*commits)
	slimit := c.QueryParam("limit")
	limit := 10
	if slimit == "all" {
		limit = commitslen
	} else if len(slimit) > 0 {
		i, err := strconv.Atoi(slimit)
		if err != nil {
			return errorResponse(c, invalidParam)
		}
		limit = i
		if i > commitslen {
			limit = commitslen
		}
	}
	return c.JSON(http.StatusOK, (*commits)[commitslen-limit:commitslen])
}

func UpdateLog(c echo.Context) error {
	ip := c.RealIP()

	log := LogUpdateRequest{}
	if err := c.Bind(&log); err != nil {
		c.Logger().Info(err)
		return c.NoContent(http.StatusBadRequest)
	}
	if err := c.Validate(&log); err != nil {
		c.Logger().Info(err)
		return c.NoContent(http.StatusBadRequest)
	}

	if err := AppendLogToStore(ip, log.Level, log.Data); err != nil {
		c.Logger().Info(err)
		return c.NoContent(http.StatusBadRequest)
	}
	return c.NoContent(http.StatusOK)
}

func UpdateHeartbeat(c echo.Context) error {
	ip := c.RealIP()
	if err := UpdateHeartbeatToStore(ip); err != nil {
		c.Logger().Info(err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func errorResponse(c echo.Context, reason errorReason) error {
	return c.JSON(errorReasonToCode[reason], map[string]string{
		"reason": string(reason),
	})
}
