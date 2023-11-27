package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)


func showHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	_helpMessage := `Commands for bot: 
	!joke   - bot sends random joke about Chuck Norris
	!join   - bot joins voice channel
	!leave  - bot leaves voice channel
	!play <URL>   - bot plays song from YOUTUBE url
	!help   - bot shows help message`

	s.ChannelMessageSend(m.ChannelID, _helpMessage)
}

func sendWorld(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "world!")
}

func sendJoke(s *discordgo.Session, m *discordgo.MessageCreate) {
	joke := getJoke()
	s.ChannelMessageSend(m.ChannelID, joke)
}

func joinVoiceChat(s *discordgo.Session, m *discordgo.MessageCreate) {
	guildID := m.GuildID
	voiceState, err := s.State.VoiceState(guildID, m.Author.ID)
	if err != nil {
		log.Println("error getting voice state: ", err)
		return
	}

	if voiceState == nil || voiceState.ChannelID == "" {
		s.ChannelMessageSend(m.ChannelID, "You need to be in a voice channel first.")
		return
	}

	voiceConnection, err = s.ChannelVoiceJoin(guildID, voiceState.ChannelID, false, true)
	if err != nil {
		log.Println("error joining voice channel: ", err)
		return
	}

	s.ChannelMessageSend(m.ChannelID, "I have joined the voice chat!")
}

func leaveVoiceChat(s *discordgo.Session, m *discordgo.MessageCreate) {
	if voiceConnection == nil {
		s.ChannelMessageSend(m.ChannelID, "I'm not in a voice channel.")
		return
	}

	err := voiceConnection.Disconnect()
	if err != nil {
		log.Println("error leaving voice channel: ", err)
		return
	}

	s.ChannelMessageSend(m.ChannelID, "I have left the voice chat!")
}