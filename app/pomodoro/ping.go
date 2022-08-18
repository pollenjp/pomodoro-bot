package pomodoro

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func pingPongMessageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.Bot {
		// interrupt conversation with "bot"
		return
	}
	log.Printf("%20s %20s > %s", message.ChannelID, message.Author.Username, message.Content)

	switch {
	case message.Content == "ping":
		if _, err := session.ChannelMessageSend(message.ChannelID, "pong"); err != nil {
			log.Print("Error sending message: ", err)
		}
	}

}
