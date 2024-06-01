package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	aa "github.com/marlonmp/bemo.go"
)

func messageCreate(session *discordgo.Session, msg *discordgo.MessageCreate) {
	if msg.Author.ID == session.State.User.ID {
		return
	}

	if msg.Content == "ping" {
		_, err := session.ChannelMessageSendReply(msg.ChannelID, "Pong!", msg.Reference())

		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

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
		fmt.Printf("[err]:cannot connect to discord: %s\n", err)
	}

	session.AddHandler(messageCreate)

	a := aa.ChannelService{}

	session.AddHandler(a.Join)
	session.AddHandler(a.Follow)

	session.Identify.Intents = discordgo.IntentsAll

	err = session.Open()

	if err != nil {
		fmt.Printf("[err]: cannot open the websocket channel: %s\n", err)
	}

	defer session.Close()

	fmt.Println("Bemo.go is running. Press ctrl-c to exit.")

	signal_chan := make(chan os.Signal, 1)

	signal.Notify(signal_chan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	<-signal_chan
}
