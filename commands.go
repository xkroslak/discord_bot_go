package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func showHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	_helpMessage := `Commands for bot: 
	!joke   - bot sends random joke about Chuck Norris
	!join   - bot joins voice channel
	!leave  - bot leaves voice channel
	!play <URL>   - bot plays song from YOUTUBE url
	!poll "QUESTION" "ANSWEAR1" "ANSWEAR2"   - bot creates poll up to 10 answears
	!drivers <YEAR>   - bot displays F1 drivers who drove in a given year 
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

func showDrivers(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	year := strings.Trim(args[0], "")
	driversJson := getDrivers(year)
	driversMessage := "Drivers on the grid in season 2023: \n"
	for _, driver := range driversJson.MRData.DriverTable.Drivers {
		driverLine := fmt.Sprintf("%s  %s %s   Number: %s   Nationality: %s \n",
			driver.Code, driver.GivenName, driver.FamilyName, driver.PermanentNumber, driver.Nationality)
		driversMessage = driversMessage + driverLine
	}
	s.ChannelMessageSend(m.ChannelID, driversMessage)
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

func makePoll(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if !checkArguments(args) {
		s.ChannelMessageSend(m.ChannelID, `Missing or wrong poll arguments: answears or question. Correct form is !poll "Question" "Answear" "Answear" ... up to 9 answears
		for example: !poll "How are you?" "Good" "Bad"`)
		return
	}

	createPoll(s, m, args)
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
