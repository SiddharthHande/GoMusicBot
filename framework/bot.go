package framework

import (
	"fmt"
	"musicbot/audio"
	"musicbot/vc"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var voiceManager *vc.VoiceManager
var Session *discordgo.Session
var queueManager *audio.QueueManager
var audioSessions *audio.AudioSessionManager

func InitBot() {

	token := os.Getenv("DISCORD_TOKEN")
	var err error
	Session, err = discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	voiceManager = vc.NewVoiceManager()
	queueManager = audio.NewQueueManager()
	audioSessions = audio.NewAudioSessionManager()

	Session.AddHandler(onMessageCreate)
	err = Session.Open()
	if err != nil {
		panic(err)
	}

	fmt.Println("Bot is running...")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	Session.Close()
}
