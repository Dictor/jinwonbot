package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
	"unicode/utf8"
)

var currentSession *discordgo.Session

func startBot(bot_token string) error {
	var err error = nil

	currentSession, err = discordgo.New("Bot " + bot_token)
	if err != nil {
		return err
	}

	currentSession.AddHandler(messageHandler)
	err = currentSession.Open()
	return err
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID { // Ignore bot's itself message
		return
	} else if strings.Contains(m.Content, "진원쿤") && utf8.RuneCountInString(m.Content) <= 5 {
		var now_answer string = "지금 바라미실은 "
		var time_string string
		if latestChangeTime == 0 {
			time_string = "(엥? 데이터가 없네요)"
		} else {
			time_string = fmt.Sprint(time.Now().Sub(time.Unix(latestChangeTime, 0)))
		}

		if currentDoorStatus {
			now_answer += "열려있습니다!, " + time_string + " 전에 열렸어요."
		} else {
			now_answer += "닫혀있습니다!, " + time_string + " 전에 닫혔어요."
		}
		s.ChannelMessageSend(m.ChannelID, now_answer)
	}
}
