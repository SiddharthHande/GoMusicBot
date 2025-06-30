package framework

import (
	commands "musicbot/cmd"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	cmd := commands.NewBotCommand(s, m, voiceManager, queueManager, audioSessions)
	args := strings.Fields(m.Content)

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
			"`!queue clear` - Clear the queue\n" +
			"`!loop one|all|off|toggle` - Set loop mode"
		s.ChannelMessageSend(m.ChannelID, helpMessage)

	case strings.HasPrefix(m.Content, "!info"):
		s.ChannelMessageSend(m.ChannelID, "🎵 This is a music bot written in Go using DiscordGo.\nSupports playback, queues, and loop modes.")

	case strings.HasPrefix(m.Content, "!join"):
		cmd.Join()

	case strings.HasPrefix(m.Content, "!play"):
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: !play <youtube_url>")
			return
		}
		cmd.Play(args[1])

	case strings.HasPrefix(m.Content, "!leave"):
		cmd.Leave()

	case strings.HasPrefix(m.Content, "!skip"):
		cmd.Skip()

	case strings.HasPrefix(m.Content, "!queue"):
		args := strings.Fields(m.Content)

		if len(args) == 2 {
			switch args[1] {
			case "clear":
				cmd.ClearQueue()
			case "shuffle":
				cmd.ShuffleQueue()
			default:
				cmd.Queue()
			}
		} else if len(args) == 3 && args[1] == "remove" {
			index, err := strconv.Atoi(args[2])
			if err != nil {
				cmd.Session.ChannelMessageSend(m.ChannelID, "⚠️ Invalid index.")
				return
			}
			cmd.RemoveFromQueue(index)
		} else if len(args) >= 4 && args[1] == "insert" {
			index, err := strconv.Atoi(args[2])
			if err != nil {
				cmd.Session.ChannelMessageSend(m.ChannelID, "⚠️ Invalid index.")
				return
			}
			url := args[3]
			cmd.InsertIntoQueue(index, url)
		} else if len(args) == 4 && args[1] == "move" {
			from, err1 := strconv.Atoi(args[2])
			to, err2 := strconv.Atoi(args[3])
			if err1 != nil || err2 != nil {
				cmd.Session.ChannelMessageSend(m.ChannelID, "⚠️ Invalid positions.")
				return
			}
			cmd.MoveInQueue(from, to)
		} else {
			cmd.Queue()
		}

	case strings.HasPrefix(m.Content, "!stop"):
		cmd.Stop()

	case strings.HasPrefix(m.Content, "!nowplaying"):
		cmd.NowPlaying()

	case strings.HasPrefix(m.Content, "!pause"):
		cmd.Pause()

	case strings.HasPrefix(m.Content, "!resume"):
		cmd.Resume()

	case strings.HasPrefix(m.Content, "!loop"):
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: !loop one | all | off | toggle")
			return
		}
		switch args[1] {
		case "one", "all", "off":
			cmd.SetLoopMode(args[1])
		case "toggle":
			cmd.ToggleLoopMode()
		default:
			s.ChannelMessageSend(m.ChannelID, "Invalid loop mode. Use: `one`, `all`, `off`, or `toggle`.")
		}

	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown command. Type `!help` for available commands.")
	}
}
