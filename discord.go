package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
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
	}

	pContent := strings.Split(m.Content, " ")
	switch len(pContent) {
	case 1:
		if strings.Contains(m.Content, "진원쿤") && utf8.RuneCountInString(m.Content) <= 5 {
			answer := "지금 바라미실은 "
			if latestChangeTime == 0 {
				if currentDoorStatus {
					answer += "열려있습니다! 언제 열렸는지는 잘 모르겠어요 ㅠㅠ"
				} else {
					answer += "닫혀있습니다! 언제 닫혔는지는 잘 모르겠어요 ㅠㅠ"
				}
			} else {
				timeString := formatSecond(int64(time.Now().Sub(time.Unix(latestChangeTime, 0)).Seconds()))
				if currentDoorStatus {
					answer += fmt.Sprintf("열려있습니다! %s전에 열렸어요!", timeString)
				} else {
					answer += fmt.Sprintf("닫혀있습니다! %s전에 닫혔어요!", timeString)
				}
			}
			log.Printf("[Message Send](%s-%s): %s → %s \n", m.Author.ID, m.Author.Username, m.Content, answer)
			s.ChannelMessageSend(m.ChannelID, answer)
		}
	case 2:
		if strings.Contains(pContent[1], "정보") {
			s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Type:  discordgo.EmbedTypeRich,
				Title: "진원쿤 정보",
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{"버전", fmt.Sprintf("`%s (%s) - %s`", gitTag, gitHash[0:6], buildDate), false},
					&discordgo.MessageEmbedField{"제작자", "25기 김정현 (kimdictor@gmail.com)", true},
					&discordgo.MessageEmbedField{"소스코드", "https://github.com/Dictor/jinwonbot", true},
					&discordgo.MessageEmbedField{"데이터 수집, 제공", "24기 주진원 (https://github.com/MainEpicenter)", true},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "바라미실에 설치된 하드웨어 (https://github.com/ibarami/IsBaramiOpen)를 통해 수집한 정보를 제공하는 웹페이지 (https://ibarami.github.io)를 크롤링하여 정보를 제공하고 있습니다.",
				},
			})
		} else if strings.Contains(pContent[1], "도움말") {
			s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Type:  discordgo.EmbedTypeRich,
				Title: "진원쿤 명령어",
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{"정보 보기", "진원쿤 정보", true},
					&discordgo.MessageEmbedField{"바라미실 문 상태 확인", "진원쿤", true},
					&discordgo.MessageEmbedField{"도움말 보기", "진원쿤 도움말", true},
				},
			})
		}
	}
}

func formatSecond(sec int64) string {
	// What is more efficient? define all variable int64 or casting everytime
	var units [4]int64                              // [second, minute, hour, day]
	var unitsRef = [4]int64{1, 60, 3600, 3600 * 24} // reference of each unit
	var unitsPostfix = [4]string{"초", "분", "시간", "일"}

	for i := 3; i >= 0; i-- {
		units[i] = sec / unitsRef[i]
		sec = sec % unitsRef[i]
	}

	var res string
	for i := 3; i >= 0; i-- {
		if units[i] != 0 {
			res += fmt.Sprintf("%d%s ", units[i], unitsPostfix[i])
		}
	}

	return res
}
