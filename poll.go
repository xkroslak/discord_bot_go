package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func checkArguments(args []string) bool {
	joinedArgs := strings.Join(args, " ")
	splittedArgs := strings.Split(joinedArgs, "\"")
	var pollArgs []string
	for i := 1; i < len(splittedArgs); i += 2 {
		pollArgs = append(pollArgs, splittedArgs[i])
	}
	if len(pollArgs) < 2 || len(pollArgs) > 10{
		return false
	}
	return true
}

func createPoll(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	joinedArgs := strings.Join(args, " ")
	splittedArgs := strings.Split(joinedArgs, "\"")
	var pollArgs []string
	for i := 1; i < len(splittedArgs); i += 2 {
		pollArgs = append(pollArgs, splittedArgs[i])
	}
	question := pollArgs[0]
	options := pollArgs[1:]
	emojis := []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣"}

	fields := make([]*discordgo.MessageEmbedField, len(options))
	for i, option := range options {
		fields[i] = &discordgo.MessageEmbedField{
			Name:  option,
			Value: emojis[i],
		}
	}

	poll := discordgo.MessageEmbed{
		Title:       "Poll",
		Description: question,
		Color:       0x00ff00,
		Fields:      fields,
	}

	message, err := s.ChannelMessageSendEmbed(m.ChannelID, &poll)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, emoji := range emojis[:len(options)] {
		s.MessageReactionAdd(m.ChannelID, message.ID, emoji)
	}

}
