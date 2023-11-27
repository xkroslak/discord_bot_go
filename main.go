package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

const (
	youtubeDL     = "yt-dlp"
	audioFolder   = "audio/"
	commandPrefix = "!"
)

var (
	voiceConnection *discordgo.VoiceConnection
	queue           []string
)

func main() {
	err := DiscordConnect()
	if err != nil {
		log.Println("FATA: Discord", err)
		return
	}
}
