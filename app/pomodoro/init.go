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
		os.Getenv("GUILD_ID"),
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

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	if err = session.Open(); err != nil {
		panic(err)
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, Info.GetGuildID(), v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	app.Destructor.Append(
		func() {

			log.Println("Removing commands...")
			for _, v := range registeredCommands {
				err := session.ApplicationCommandDelete(session.State.User.ID, Info.GetGuildID(), v.ID)
				if err != nil {
					log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
				}
			}

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
