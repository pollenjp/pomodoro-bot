package pomodoro

import "log"

var (
	Info info
)

type info struct {
	guildID                  GuildID
	channelIDForNotification ChannelID
	channelIDForPomodoroVC   ChannelID
}

func InitInfo(
	guildID string,
	channelIDForNotification string,
	channelIDForPomodoroVC string,
) {
	Info = info{
		guildID:                  guildID,
		channelIDForNotification: channelIDForNotification,
		channelIDForPomodoroVC:   channelIDForPomodoroVC,
	}
}

func (i *info) GetGuildID() GuildID {
	if len(i.guildID) == 0 {
		log.Fatal("no guild id exists.")
	}
	return i.guildID
}

func (i *info) GetChannelIDForNotification() ChannelID {
	if len(i.channelIDForNotification) == 0 {
		log.Fatal("no channel id for notification exists.")
	}
	return i.channelIDForNotification
}

func (i *info) GetChannelIDForPomodoroVC() ChannelID {
	if len(i.channelIDForPomodoroVC) == 0 {
		log.Fatal("no channel id for notification exists.")
	}
	return i.channelIDForPomodoroVC
}
