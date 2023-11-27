package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

func playSongVoiceChat(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if voiceConnection == nil {
		s.ChannelMessageSend(m.ChannelID, "I'm not in a voice channel.")
		return
	}

	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "You need to provide a YouTube URL to play.")
		return
	}

	youtubeURL := args[0]

	if !strings.Contains(youtubeURL, "youtube.com") && !strings.Contains(youtubeURL, "youtu.be") {
		s.ChannelMessageSend(m.ChannelID, "That's not a valid YouTube URL.")
		return
	}

	cmd := exec.Command("yt-dlp", "-f", "bestaudio", "-g", youtubeURL)
	output, err := cmd.Output()
	if err != nil {
		log.Println("error getting audio stream URL: ", err)
		return
	}

	audioStreamURL := strings.TrimSpace(string(output))

	s.ChannelMessageSend(m.ChannelID, "I'm playing the song!")

	err = voiceConnection.Speaking(true)
	if err != nil {
		log.Println("error setting speaking: ", err)
		return
	}

	// TODO: Create a new ffmpeg command to stream the audio
	ffmpeg := exec.Command("ffmpeg", "-i", audioStreamURL, "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")
	//ffmpeg := exec.Command("ffmpeg", "-i", audioStreamURL, "-c:a", "aac", "-b:a", "128k", "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")
	ffmpegOut, err := ffmpeg.StdoutPipe()
	if err != nil {
		log.Println("error creating ffmpeg stdout pipe: ", err)
		return
	}

	err = ffmpeg.Start()
	if err != nil {
		log.Println("error starting ffmpeg: ", err)
		return
	}

	// TODO: Create a new discordgo stream to play the audio
	stream := NewStream(voiceConnection, ffmpegOut)

	// TODO: Start the stream and wait for it to finish
	stream.Play()
	stream.Wait()

	err = ffmpeg.Process.Kill()
	if err != nil {
		log.Println("error stopping ffmpeg: ", err)
		return
	}

	err = voiceConnection.Speaking(false)
	if err != nil {
		log.Println("error setting speaking: ", err)
		return
	}
}

//type Stream struct {
//	voiceConnection *discordgo.VoiceConnection
//	ffmpegOut       io.Reader
//	volume          float64
//}

type Stream struct {
	voiceConnection *discordgo.VoiceConnection
	ffmpegOut       io.ReadCloser
	volume          float64
}

func NewStream(vc *discordgo.VoiceConnection, r io.ReadCloser) *Stream {
	s := &Stream{
		voiceConnection: vc,
		ffmpegOut:       r,
		volume:          0.5,
	}

	return s
}

func (s *Stream) Play() {
	s.voiceConnection.Speaking(true)

	defer s.voiceConnection.Speaking(false)

	for {
		buf := make([]byte, 960)
		_, err := s.ffmpegOut.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println("error reading ffmpeg output: ", err)
			}
			break
		}

		s.voiceConnection.OpusSend <- buf
	}
}

func (s *Stream) Wait() {
	for {
		if len(s.voiceConnection.OpusSend) == 0 {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}
}


