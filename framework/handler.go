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

	if len(args) == 0 || !strings.HasPrefix(args[0], "!") {
		return
	}

	switch args[0] {
	case "!ping":
		s.ChannelMessageSend(m.ChannelID, "Pong!")

	case "!help":
		s.ChannelMessageSend(m.ChannelID,
			"Available commands:\n"+
				"`!ping` - Responds with Pong!\n"+
				"`!help` - Displays this help message\n"+
				"`!join`, `!leave` - Voice connection\n"+
				"`!play <url>` - Play a YouTube video or playlist\n"+
				"`!pause`, `!resume`, `!skip`, `!stop`\n"+
				"`!queue` - Show queue\n"+
				"`!queue clear|shuffle` - Manage queue\n"+
				"`!queue insert <index> <url>`\n"+
				"`!queue remove <index>`\n"+
				"`!queue move <from> <to>`\n"+
				"`!loop one|all|off|toggle` - Set loop mode\n"+
				"`!nowplaying`, `!search <query>`")

	case "!info":
		s.ChannelMessageSend(m.ChannelID, "🎵 This is a music bot written in Go using DiscordGo.\nSupports playback, queues, and loop modes.")

	case "!join":
		cmd.Join()

	case "!play":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: `!play <youtube_url>`")
			return
		}
		cmd.Play(args[1])

	case "!leave":
		cmd.Leave()

	case "!skip":
		cmd.Skip()

	case "!queue":
		switch {
		case len(args) == 2 && args[1] == "clear":
			cmd.ClearQueue()
		case len(args) == 2 && args[1] == "shuffle":
			cmd.ShuffleQueue()
		case len(args) == 3 && args[1] == "remove":
			index, err := strconv.Atoi(args[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "⚠️ Invalid index.")
				return
			}
			cmd.RemoveFromQueue(index)
		case len(args) >= 4 && args[1] == "insert":
			index, err := strconv.Atoi(args[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "⚠️ Invalid index.")
				return
			}
			url := args[3]
			cmd.InsertIntoQueue(index, url)
		case len(args) == 4 && args[1] == "move":
			from, err1 := strconv.Atoi(args[2])
			to, err2 := strconv.Atoi(args[3])
			if err1 != nil || err2 != nil {
				s.ChannelMessageSend(m.ChannelID, "⚠️ Invalid move positions.")
				return
			}
			cmd.MoveInQueue(from, to)
		default:
			cmd.Queue()
		}

	case "!stop":
		cmd.Stop()

	case "!nowplaying":
		cmd.NowPlaying()

	case "!pause":
		cmd.Pause()

	case "!resume":
		cmd.Resume()

	case "!loop":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: `!loop one | all | off | toggle`")
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

	case "!search":
		query := strings.TrimSpace(strings.Join(args[1:], " "))
		if query == "" {
			s.ChannelMessageSend(m.ChannelID, "Usage: `!search <query>`")
			return
		}
		go cmd.Search(query)

	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown command. Type `!help` for available commands.")
	}
}
