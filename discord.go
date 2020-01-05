package main

import (
	"fmt"
	"github.com/andersfylling/disgord"
)

func getMessage(session disgord.Session, evt *disgord.MessageCreate) {
	msg := evt.Message
	fmt.Println(msg.Author.String() + ": " + msg.Content) // Anders#7248{435358734985}: Hello there
}
