package main

import (
	"encoding/json"
	"fmt"
	"github.com/dictor/justlog"
	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters/sqlite"
	"io/ioutil"
	"log"
	"time"
)

type DoorStatus struct {
	Time   int64 `db:"timestamp,key"`
	Status bool  `db:"status"`
}

type TimeBlock struct {
	Day     string `json:"day"`  // "YYMMDD"
	Hour    int    `json:"hour"` // integer hour
	Minute  string `json:"min"`  // 0~9="0", 10~19="1", 20~29="2", 30~39="3", 40~49="4", 50~59="5"
	RawUnix int64  `json:"timestamp"`
}

func main() {
	// log initializing
	justlog.MustStream(justlog.SetStream(justlog.MustPath(justlog.SetPath())))

	// select all rows from db
	db, err := godb.Open(sqlite.Adapter, "input.db")
	defer db.Close()
	res := make([]DoorStatus, 0)
	err = db.Select(&res).OrderBy("timestamp ASC").Do()
	if err != nil {
		log.Panic(err)
	}

	// create result blocks
	log.Printf("%d rows retireved.", len(res))
	timeblks := make([]TimeBlock, 0)
	for i := 1; i < len(res); i++ {
		timeblks = append(timeblks, newBlock(res[i]))
	}

	// encoding blocks to json
	jsonres, err := json.Marshal(timeblks)
	if err != nil {
		log.Panic(err)
	}

	ioutil.WriteFile("result.json", jsonres, 0775)
	log.Printf("%d bytes writed, task complete.", len(jsonres))
}

func newBlock(d DoorStatus) TimeBlock {
	t := time.Unix(d.Time, 0)
	m := fmt.Sprintf("%02d", t.Minute())
	res := TimeBlock{fmt.Sprintf("%02d%02d%02d", t.Year()-2000, t.Month(), t.Day()), t.Hour(), m[0:1], d.Time}
	return res
}
