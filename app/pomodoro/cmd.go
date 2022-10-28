package pomodoro

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "pomodoro",
			Description: "pomodoro commands",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "ping",
					Description: "ping-pong",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "start",
					Description: "start pomodoro",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "stop",
					Description: "stop pomodoro",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"pomodoro": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("pomodoro command: %+v", i.ApplicationCommandData())
			options := i.ApplicationCommandData().Options
			content := ""

			switch options[0].Name {
			case "ping":
				content = "pong!"

			case "start":
				user := i.Member.User
				content = fmt.Sprintf("Hi %s! See <#%s>!", user.Username, Info.GetChannelIDForNotification())

				// user が VC にいるかどうかを確認する
				var voiceState *discordgo.VoiceState
				var err error
				if voiceState, err = s.State.VoiceState(Info.GetGuildID(), user.ID); err != nil {
					log.Printf("Failed to get %s's voice state: %v", user.Username, err)
				}
				if voiceState == nil {
					log.Printf("User is not in voice channel")
					content += "\n"
					content += "You are not in voice channel."
					break
				}

				log.Printf("User is in voice channel: %v", voiceState.ChannelID)

				// VC にいる場合は pomodoro を開始する
				if pomodoro, err := getPomodoroWithLock(s, i.GuildID, Info.GetChannelIDForNotification()); err != nil {
					return
				} else {
					defer unlockPomodoro(Info.GetGuildID())
					pomodoro.AddUser(*user)
				}
			case "stop":
				user := i.Member.User
				content = fmt.Sprintf("Bye %s! See <#%s>!", user.Username, Info.GetChannelIDForNotification())

				// user が VC にいるかどうかを確認する
				var voiceState *discordgo.VoiceState
				var err error
				if voiceState, err = s.State.VoiceState(Info.GetGuildID(), user.ID); err != nil {
					log.Printf("Failed to get %s's voice state: %v", user.Username, err)
				}
				if voiceState == nil {
					log.Printf("User is not in voice channel")
					// いなければ忠告
					content += "\n"
					content += fmt.Sprintf("<@%s>! You should run `/pomodoro stop` command in some voice channel.", user.ID)
					content += "\n"
					content += "Otherwise, your server mute and deaf status may not be released."
				}

				// pomodoro を停止
				if pomodoro, err := getPomodoroWithLock(s, i.GuildID, Info.GetChannelIDForNotification()); err != nil {
					return
				} else {
					defer releaseOrUnlockPomodoro(pomodoro, i.GuildID)
					pomodoro.RemoveMember(user.ID)
				}
			default:
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
		},
	}
)
