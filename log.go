package main

import (
	"github.com/kardianos/osext"
	"io"
	"log"
	"os"
	"time"
)

func setLogPath() (string, error) {
	var path string
	path, err := osext.ExecutableFolder()
	path += "/log"
	prepareDirectory(path)
	return path, err
}

func setLogStream(path string) (*os.File, error) {
	fpLog, err := os.OpenFile(path+"/"+time.Now().Format("2006-01-02T15_04_05")+".txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0775)
	if err != nil {
		return nil, err
	} else {
		multiWriter := io.MultiWriter(fpLog, os.Stdout)
		log.SetOutput(multiWriter)
	}
	return fpLog, nil
}

func prepareDirectory(dir ...string) {
	for _, val := range dir {
		if _, err := os.Stat(val); os.IsNotExist(err) {
			os.Mkdir(val, 0775)
		}
	}
}
