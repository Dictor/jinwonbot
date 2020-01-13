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
			if currentDoorStatus {
				now_answer += "열려있습니다! 언제 열렸는지는 잘 모르겠어요 ㅠㅠ"
			} else {
				now_answer += "닫혀있습니다! 언제 닫혔는지는 잘 모르겠어요 ㅠㅠ"
			}
		} else {
			time_string = formatSecond(int64(time.Now().Sub(time.Unix(latestChangeTime, 0)).Seconds()))
			if currentDoorStatus {
				now_answer += fmt.Sprintf("열려있습니다! %s전에 열렸어요!", time_string)
			} else {
				now_answer += fmt.Sprintf("닫혀있습니다! %s전에 닫혔어요!", time_string)
			}
		}

		s.ChannelMessageSend(m.ChannelID, now_answer)
	}
}

func formatSecond(sec int64) string {
	// What is more efficient? define all variable int64 or casting everytime
	var units [4]int64                               // [second, minute, hour, day]
	var units_ref = [4]int64{1, 60, 3600, 3600 * 24} // reference of each unit
	var units_postfix = [4]string{"초", "분", "시간", "일"}

	for i := 3; i >= 0; i-- {
		units[i] = sec / units_ref[i]
		sec = sec % units_ref[i]
	}

	var res string
	for i := 3; i >= 0; i-- {
		if units[i] != 0 {
			res += fmt.Sprintf("%d%s ", units[i], units_postfix[i])
		}
	}

	return res
}
