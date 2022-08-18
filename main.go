package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pollenjp/pomodoro-bot/app"
	_ "github.com/pollenjp/pomodoro-bot/app/pomodoro"
)

func init() {
}

func main() {
	defer app.Destructor.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	log.Print("booted!!!")

	<-sc

}
