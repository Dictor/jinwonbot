package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
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
	if strings.Contains(m.Content, "진원쿤") {
		GlobalLogger.WithFields(logrus.Fields{
			"author_id":   m.Author.ID,
			"author_name": m.Author.Username,
			"content":     m.Content,
		}).Infoln("bot is called")

		callCount := GetInfo(CallCount)
		if callCount == "" {
			if err := SetInfoToStore(CallCount, "1"); err != nil {
				GlobalLogger.WithError(err).Error("fail to set call count as 1")
			}
		} else {
			n, err := strconv.ParseInt(callCount, 10, 64)
			if err == nil {
				n++
				err = SetInfoToStore(CallCount, strconv.FormatInt(n, 10))
				if err != nil {
					GlobalLogger.WithError(err).Error("fail to increase call count")
				}
			} else {
				GlobalLogger.WithError(err).Error("fail to parse call count")
			}
		}
		SaveStore()
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
			logSendResult(s.ChannelMessageSend(m.ChannelID, answer))
		}
	case 2:
		if !strings.Contains(pContent[0], "진원쿤") {
			return
		}
		if strings.Contains(pContent[1], "정보") {
			logSendResult(s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Type:  discordgo.EmbedTypeRich,
				Title: "진원쿤 정보",
				Fields: []*discordgo.MessageEmbedField{
					{Name: "버전", Value: fmt.Sprintf("`%s (%s) - %s`", gitTag, gitHash[0:6], buildDate), Inline: false},
					{Name: "제작자", Value: "[25기 김정현](https://github.com/Dictor)", Inline: true},
					{Name: "소스코드", Value: "[깃헙 저장소](https://github.com/Dictor/jinwonbot)", Inline: true},
					{Name: "데이터 수집, 제공", Value: "[24기 주진원](https://github.com/MainEpicenter)", Inline: true},
					{Name: "인기만점진원쿤", Value: fmt.Sprintf("지금까지 %s번 불렸어요!", GetInfo(CallCount)), Inline: true},
				},
				Description: "바라미실에 설치된 [하드웨어](https://github.com/ibarami/IsBaramiOpen)를 통해 수집한 정보를 제공하는 [웹페이지](https://ibarami.github.io)를 크롤링하여 정보를 제공하고 있습니다.",
			}))
		} else if strings.Contains(pContent[1], "도움말") {
			logSendResult(s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Type:  discordgo.EmbedTypeRich,
				Title: "진원쿤 명령어",
				Fields: []*discordgo.MessageEmbedField{
					{Name: "정보 보기", Value: "`진원쿤 정보`", Inline: true},
					{Name: "문 상태 보기", Value: "`진원쿤`", Inline: true},
					{Name: "도움말 보기", Value: "`진원쿤 도움말`", Inline: true},
					{Name: "개발자 정보 보기", Value: "`진원쿤 디버그`", Inline: true},
				},
				Description: "원하는 기능의 명령어를 채팅방에 입력하면 됩니다.",
			}))
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

			logSendResult(s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Type:  discordgo.EmbedTypeRich,
				Title: "진원쿤 디버그 정보",
				Fields: []*discordgo.MessageEmbedField{
					{Name: "총 로드한 커밋 수", Value: fmt.Sprintf("%d개", len(*commits)), Inline: true},
					{Name: "마지막 개문 기록", Value: openRecord, Inline: false},
					{Name: "마지막 폐문 기록", Value: closeRecord, Inline: false},
					{Name: "최근 5개 기록", Value: recentCommits, Inline: false},
					{Name: "Store 정보", Value: fmt.Sprintf("버전 %d, %s", GetStoreVersion(), GetStoreDebugString()), Inline: false},
					{Name: "하트비트 리스트", Value: GetHeartbeatString(), Inline: false},
				},
			}))
			for ip, v := range GetLogString() {
				logSendResult(s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
					Type:  discordgo.EmbedTypeRich,
					Title: "진원쿤 디버그 정보 (로그)",
					Fields: []*discordgo.MessageEmbedField{
						{Name: fmt.Sprintf("%s의 로그", ip), Value: v, Inline: false},
					},
				}))
			}
		} else if strings.Contains(pContent[1], "윤성") {
			if len(ysArt) == 0 {
				logSendResult(s.ChannelMessageSend(m.ChannelID, "엥? 뭔가 문제가 있는데요..."))
				return
			}
			logSendResult(s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Type:        discordgo.EmbedTypeArticle,
				Description: ysArt,
			},
			))
		}
	}
}

func logSendResult(msg *discordgo.Message, err error) {
	if err != nil {
		GlobalLogger.WithError(err).Errorln("fail to send message")
		AppendLogToStore("host", "E", "logSendResult negative")
		return
	}
	GlobalLogger.WithFields(logrus.Fields{
		"content": msg.Content,
		"channel": msg.ChannelID,
	}).Infoln("message is sent")
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
