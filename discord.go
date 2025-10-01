package main

import (
	"fmt"
	"math"
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
			latestHeartbeat := GetLatestHeartbeat()
			if latestHeartbeat == nil {
				answer += "\n[경고] 수신된 하트비트가 없습니다. 유효하지 않은 과거 데이터가 표시될 수 있습니다."
			} else if latestHeartbeat.int64 > 3600 {
				answer += "\n[경고] 주 단말기의 마지막 하트비트가 수신된지 1시간이 지났습니다. 주 단말기가 고장나 올바른 정보를 표시하지 않을 수도 있습니다."
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

			latestHeartbeat := GetLatestHeartbeat()
			latestHeartbeatString := ""
			if latestHeartbeat == nil {
				latestHeartbeatString = "수신된 하트비트가 없습니다."
			} else {
				latestHeartbeatString = fmt.Sprintf("%s 전에 하트비트를 발신한 %s가 주 단말기입니다.", formatSecond(latestHeartbeat.int64), latestHeartbeat.string)
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
					{Name: "하트비트 리스트", Value: fmt.Sprintf("%s\n%s", GetHeartbeatString(), latestHeartbeatString), Inline: false},
				},
			}))

			logSizeInfo := []*discordgo.MessageEmbedField{
				{Name: "로그를 확인하기 위해 '진원쿤 디버그로그 <단말기 주소> <페이지 번호> 명령어를 사용하세요.'", Value: "", Inline: false},
			}
			for ip, v := range GetLogString() {
				logSizeInfo = append(logSizeInfo, &discordgo.MessageEmbedField{Name: fmt.Sprintf("단말기 '%s'", ip), Value: fmt.Sprintf("%d bytes", len(v)), Inline: false})
			}
			logSendResult(s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Type:   discordgo.EmbedTypeRich,
				Title:  "로그를 기록한 단말기 목록",
				Fields: logSizeInfo,
			}))
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
	case 4:
		if pContent[1] == "디버그로그" {
			logs := GetLogString()
			targetLog, exist := logs[pContent[2]]
			const bytesPerPage int = 2000

			if !exist {
				logSendResult(s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("'%s' 단말기가 존재하지 않습니다.", pContent[2])))
				return
			}

			targetLogPageCount := int(math.Ceil(float64(len(targetLog)) / float64(bytesPerPage)))
			if targetLogPageCount <= 0 {
				logSendResult(s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("'%s' 단말기가 존재하지만, 기록된 로그가 없습니다.", pContent[2])))
				return
			}

			targetPage, err := strconv.Atoi(pContent[3])
			if err != nil {
				logSendResult(s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("'%s'는 올바른 페이지 번호가 아닙니다. (%s)", pContent[3], err)))
				return
			}
			if targetPage <= 0 {
				logSendResult(s.ChannelMessageSend(m.ChannelID, "페이지 번호는 음수일 수 없습니다."))
				return
			}
			if targetPage > targetLogPageCount {
				logSendResult(s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("'%s' 단말기의 로그는 %d 페이지까지만 존재합니다.", pContent[2], targetLogPageCount)))
				return
			}

			startIndex := bytesPerPage * (targetPage - 1)
			endIndex := min(len(targetLog), bytesPerPage*targetPage)
			targetLogPage := targetLog[startIndex:endIndex]
			logSendResult(s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Type:  discordgo.EmbedTypeRich,
				Title: fmt.Sprintf("진원쿤 단말기 '%s'의 로그 (%d 페이지)", pContent[2], targetPage),
				Fields: []*discordgo.MessageEmbedField{
					{Name: "로그", Value: targetLogPage, Inline: false},
				},
			}))
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
