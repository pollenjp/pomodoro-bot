package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pollenjp/pomodoro-bot/app"
	_ "github.com/pollenjp/pomodoro-bot/app/pomodoro"
)

func init() {
	setTimezone()
}

func main() {
	defer app.Destructor.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	log.Print("booted!!!")

	<-sc

}

func setTimezone() {
	location := os.Getenv("TZ")
	default_offset := 9 * 60 * 60
	if len(location) == 0 {
		location = "Asia/Tokyo"
	}
	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.FixedZone(location, default_offset)
	}
	time.Local = loc
}
