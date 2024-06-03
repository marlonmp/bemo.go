package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/marlonmp/bemogo"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	token := fmt.Sprintf("Bot %s", os.Getenv("DISCORD_BOT_TOKEN"))

	session, err := discordgo.New(token)

	if err != nil {
		fmt.Printf("cannot connect to discord: %s\n", err)
	}

	session.Identify.Intents = discordgo.IntentsAll

	service := bemogo.ChannelService{}

	session.AddHandler(service.AddCommands)

	err = session.Open()

	if err != nil {
		fmt.Printf("cannot open the websocket channel: %s\n", err)
	}

	defer session.Close()

	fmt.Println("bemogo is running. Press ctrl-c to exit.")

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	<-signalChan

	service.RemoveCommands(session)
}
