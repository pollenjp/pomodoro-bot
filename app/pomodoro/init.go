package pomodoro

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/pollenjp/pomodoro-bot/app"
)

func init() {
	InitInfo(
		os.Getenv("CHANNEL_ID_FOR_NOTIFICATION"),
		os.Getenv("CHANNEL_ID_FOR_POMODORO_VC"),
	)

	discordToken := loadToken()

	fmt.Printf("Info: %+v\n", Info)

	session, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal("Error in create session")
	}

	session.AddHandler(pingPongMessageHandler)
	session.AddHandler(onVoiceStateUpdate)

	if err = session.Open(); err != nil {
		panic(err)
	}

	app.Destructor.Append(
		func() {
			session.Close()
		},
	)

	log.Print("bot is running...")

}

func loadToken() string {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("no discord token exists.")
	}
	return token
}
