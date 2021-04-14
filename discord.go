package main

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

var currentSession *discordgo.Session

func startBot(botToken string) error {
	var err error = nil

	currentSession, err = discordgo.New("Bot " + botToken)
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
			GlobalLogger.Infof("[Message Send](%s-%s): %s → %s \n", m.Author.ID, m.Author.Username, m.Content, answer)
			s.ChannelMessageSend(m.ChannelID, answer)
		}
	case 2:
		if !strings.Contains(pContent[0], "진원쿤") {
			return
		}
		if strings.Contains(pContent[1], "정보") {
			s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Type:  discordgo.EmbedTypeRich,
				Title: "진원쿤 정보",
				Fields: []*discordgo.MessageEmbedField{
					{Name: "버전", Value: fmt.Sprintf("`%s (%s) - %s`", gitTag, gitHash[0:6], buildDate), Inline: false},
					{Name: "제작자", Value: "[25기 김정현](https://github.com/Dictor)", Inline: true},
					{Name: "소스코드", Value: "[깃헙 저장소](https://github.com/Dictor/jinwonbot)", Inline: true},
					{Name: "데이터 수집, 제공", Value: "[24기 주진원](https://github.com/MainEpicenter)", Inline: true},
				},
				Description: "바라미실에 설치된 [하드웨어](https://github.com/ibarami/IsBaramiOpen)를 통해 수집한 정보를 제공하는 [웹페이지](https://ibarami.github.io)를 크롤링하여 정보를 제공하고 있습니다.",
			})
		} else if strings.Contains(pContent[1], "도움말") {
			s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Type:  discordgo.EmbedTypeRich,
				Title: "진원쿤 명령어",
				Fields: []*discordgo.MessageEmbedField{
					{Name: "정보 보기", Value: "`진원쿤 정보`", Inline: true},
					{Name: "문 상태 보기", Value: "`진원쿤`", Inline: true},
					{Name: "도움말 보기", Value: "`진원쿤 도움말`", Inline: true},
					{Name: "개발자 정보 보기", Value: "`진원쿤 디버그`", Inline: true},
				},
				Description: "원하는 기능의 명령어를 채팅방에 입력하면 됩니다.",
			})
		} else if strings.Contains(pContent[1], "디버그") {
			commits := GetAllCommits()
			var openRecord, closeRecord, recentCommits string

			if latestOpen, err := SelectLatestStatus(true); err != nil {
				openRecord = fmt.Sprintf("오류: %s", err)
			} else {
				openRecord = fmt.Sprint(latestOpen)
			}
			if latestClose, err := SelectLatestStatus(false); err != nil {
				closeRecord = fmt.Sprintf("오류: %s", err)
			} else {
				closeRecord = fmt.Sprint(latestClose)
			}

			for _, c := range (*commits)[len(*commits)-5 : len(*commits)] {
				recentCommits += fmt.Sprintf("- %s\n", c)
			}

			s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Type:  discordgo.EmbedTypeRich,
				Title: "진원쿤 디버그 정보",
				Fields: []*discordgo.MessageEmbedField{
					{Name: "총 로드한 커밋 수", Value: fmt.Sprintf("%d개", len(*commits)), Inline: true},
					{Name: "마지막 개문 기록", Value: openRecord, Inline: false},
					{Name: "마지막 폐문 기록", Value: closeRecord, Inline: false},
					{Name: "최근 5개 기록", Value: recentCommits, Inline: false},
				},
			})
		} else if strings.Contains(pContent[1], "윤성") {
			count := 0
			if _, err := fmt.Sscanf(pContent[1], "윤성%d", &count); err != nil {
				s.ChannelMessageSend(m.ChannelID, "제가 알아들을 수 없는 값인 것 같아요 ㅠㅠ")
				return
			}
			if count > 10 {
				s.ChannelMessageSend(m.ChannelID, "뇌절 멈춰!")
				return
			}
			for i := 1; i <= 10; i++ {
				s.ChannelMessageSend(m.ChannelID, "!현석 윤성")
			}
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
