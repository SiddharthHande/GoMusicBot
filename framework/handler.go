package framework

import (
	commands "musicbot/cmd"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	cmd := commands.NewBotCommand(s, m, voiceManager, queueManager, audioSessions)
	switch {
	case strings.HasPrefix(m.Content, "!ping"):
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	case strings.HasPrefix(m.Content, "!help"):
		helpMessage := "Available commands:\n" +
			"`!ping` - Responds with Pong!\n" +
			"`!help` - Displays this help message"
		s.ChannelMessageSend(m.ChannelID, helpMessage)
	case strings.HasPrefix(m.Content, "!info"):
		infoMessage := "This is a simple music bot written in Go using the DiscordGo library.\n" +
			"It can respond to basic commands like `!ping` and `!help`."
		s.ChannelMessageSend(m.ChannelID, infoMessage)
	case strings.HasPrefix(m.Content, "!join"):
		cmd.Join()
	case strings.HasPrefix(m.Content, "!play"):
		args := strings.Fields(cmd.Message.Content)
		if len(args) < 2 {
			cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "Usage: !play <youtube_url>")
			return
		}
		cmd.Play(args[1])
	case strings.HasPrefix(m.Content, "!leave"):
		cmd.Leave()
	case strings.HasPrefix(m.Content, "!skip"):
		cmd.Skip()
	case strings.HasPrefix(m.Content, "!queue"):
		cmd.Queue()
	case strings.HasPrefix(m.Content, "!stop"):
		cmd.Stop()
	case strings.HasPrefix(m.Content, "!nowplaying"):
		cmd.NowPlaying()
	case strings.HasPrefix(m.Content, "!pause"):
		cmd.Pause()
	case strings.HasPrefix(m.Content, "!resume"):
		cmd.Resume()
	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown command. Type `!help` for available commands.")
	}
}
