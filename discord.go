package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func DiscordConnect() (err error) {
	sess, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		return err
	}

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if m.Content == "hello" {
			s.ChannelMessageSend(m.ChannelID, "world!")
		}
	})
	sess.AddHandler(messageCreate)

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer sess.Close()

	fmt.Println("The bot is online! You can end it using CTRL+C")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, commandPrefix) {
		args := strings.Split(m.Content, " ")
		command := args[0][len(commandPrefix):]

		switch command {
		case "hello":
			sendWorld(s, m)
		case "help":
			showHelp(s, m)
		case "poll":
			makePoll(s, m, args[1:])
		case "joke":
			sendJoke(s, m)
		case "drivers":
			// TODO: input check
			showDrivers(s, m, args[1:])
		case "join":
			joinVoiceChat(s, m)
		case "leave":
			leaveVoiceChat(s, m)
		case "play":
			playSongVoiceChat(s, m, args[1:])
		default:
			s.ChannelMessageSend(m.ChannelID, "Sorry, I don't recognize that command.")
		}
	}
}
