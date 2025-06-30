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
			"`!help` - Displays this help message\n" +
			"`!join` - Join your voice channel\n" +
			"`!play <url>` - Play a YouTube video or playlist\n" +
			"`!pause`, `!resume`, `!skip`, `!stop`, `!leave`\n" +
			"`!queue`, `!nowplaying`\n" +
			"`!loop one|all|off|toggle` - Set loop mode"
		s.ChannelMessageSend(m.ChannelID, helpMessage)

	case strings.HasPrefix(m.Content, "!info"):
		infoMessage := "🎵 This is a music bot written in Go using DiscordGo.\nSupports playback, queues, and loop modes."
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

	case strings.HasPrefix(m.Content, "!loop"):
		args := strings.Fields(cmd.Message.Content)
		if len(args) < 2 {
			cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "Usage: !loop one | all | off | toggle")
			return
		}
		switch args[1] {
		case "one", "all", "off":
			cmd.SetLoopMode(args[1])
		case "toggle":
			cmd.ToggleLoopMode()
		default:
			cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "Invalid loop mode. Use: `one`, `all`, `off`, or `toggle`.")
		}

	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown command. Type `!help` for available commands.")
	}
}
