package pomodoro

import "log"

var (
	Info info
)

type info struct {
	channelIDForNotification ChannelID
	channelIDForPomodoroVC   ChannelID
}

func InitInfo(
	channelIDForNotification string,
	channelIDForPomodoroVC string,
) {
	Info = info{
		channelIDForNotification: channelIDForNotification,
		channelIDForPomodoroVC:   channelIDForPomodoroVC,
	}
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
