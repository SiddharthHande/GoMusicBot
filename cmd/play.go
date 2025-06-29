package commands

import (
	"fmt"
	"musicbot/audio"
	"musicbot/vc"

	"github.com/bwmarrin/discordgo"
)

type BotCommand struct {
	Session       *discordgo.Session
	Message       *discordgo.MessageCreate
	VoiceManager  *vc.VoiceManager
	QueueManager  *audio.QueueManager
	AudioSessions *audio.AudioSessionManager // ✅ Thread-safe session manager
}

func NewBotCommand(s *discordgo.Session, m *discordgo.MessageCreate, vc *vc.VoiceManager, queue *audio.QueueManager, sessions *audio.AudioSessionManager) *BotCommand {
	return &BotCommand{
		Session:       s,
		Message:       m,
		VoiceManager:  vc,
		QueueManager:  queue,
		AudioSessions: sessions,
	}
}

func (cmd *BotCommand) getUserVoiceChannelID() string {
	guild, _ := cmd.Session.State.Guild(cmd.Message.GuildID)
	if guild == nil {
		return ""
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == cmd.Message.Author.ID {
			return vs.ChannelID
		}
	}
	return ""
}

func (cmd *BotCommand) Join() {
	guildID := cmd.Message.GuildID
	userChannelID := cmd.getUserVoiceChannelID()

	if userChannelID == "" {
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "You must be in a voice channel.")
		return
	}

	_, err := cmd.VoiceManager.Join(cmd.Session, guildID, userChannelID)
	if err != nil {
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "Failed to join VC: "+err.Error())
		return
	}

	cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "Joined your voice channel.")
}

func (cmd *BotCommand) Leave() {
	guildID := cmd.Message.GuildID

	if state, ok := cmd.AudioSessions.Get(guildID); ok {
		state.Conn.Stop()
		cmd.AudioSessions.Delete(guildID)
	}

	queue := cmd.QueueManager.Get(guildID)
	queue.Clear()
	queue.CurrentTrack = nil
	queue.IsPlaying = false

	err := cmd.VoiceManager.Leave(guildID)
	if err != nil {
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "⚠️ Failed to leave VC: "+err.Error())
		return
	}

	cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "👋 Disconnected from voice channel.")
}

func (cmd *BotCommand) Play(youtubeURL string) {
	guildID := cmd.Message.GuildID
	userChannelID := cmd.getUserVoiceChannelID()

	if userChannelID == "" {
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "You must be in a voice channel.")
		return
	}

	vc, ok := cmd.VoiceManager.Get(guildID)
	if !ok {
		cmd.Join()
		vc, ok = cmd.VoiceManager.Get(guildID)
		if !ok {
			cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "Failed to get voice connection.")
			return
		}
	}

	queue := cmd.QueueManager.Get(guildID)
	queue.Enqueue(&audio.Track{URL: youtubeURL})

	if _, exists := cmd.AudioSessions.Get(guildID); !exists {
		cmd.AudioSessions.Set(guildID, &audio.GuildAudioState{Conn: audio.NewConnection(vc)})
	}

	if queue.IsPlaying {
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "✅ Added to queue.")
		return
	}

	go cmd.startQueuePlayback(guildID, vc, queue)
}

func (cmd *BotCommand) startQueuePlayback(guildID string, vc *discordgo.VoiceConnection, queue *audio.Queue) {
	queue.IsPlaying = true
	defer func() {
		queue.IsPlaying = false
		queue.CurrentTrack = nil
		_ = cmd.VoiceManager.Leave(guildID)
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "👋 Finished playback. Left voice channel.")
	}()

	for {
		track := queue.Dequeue()
		if track == nil {
			break
		}
		queue.CurrentTrack = track

		state := &audio.GuildAudioState{Conn: audio.NewConnection(vc)}
		cmd.AudioSessions.Set(guildID, state)
		audioConn := state.Conn

		err := audioConn.Play(track.URL, &state.Paused, &state.Mutex)
		if err != nil {
			cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "⚠️ Error playing track: "+err.Error())
			continue
		}
	}
}

func (cmd *BotCommand) Stop() {
	guildID := cmd.Message.GuildID
	queue := cmd.QueueManager.Get(guildID)

	queue.Clear()
	queue.CurrentTrack = nil
	queue.IsPlaying = false

	if state, ok := cmd.AudioSessions.Get(guildID); ok {
		state.Conn.Stop()
		cmd.AudioSessions.Delete(guildID)
	}

	cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "⏹️ Stopped playback and cleared the queue.")
}

func (cmd *BotCommand) Skip() {
	guildID := cmd.Message.GuildID
	queue := cmd.QueueManager.Get(guildID)
	state, ok := cmd.AudioSessions.Get(guildID)
	if !ok || state.Conn == nil {
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "❌ Nothing is currently playing.")
		return
	}

	queue.CurrentTrack = nil
	state.Conn.Stop()

	cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "⏭️ Skipped current track.")
}

func (cmd *BotCommand) Queue() {
	guildID := cmd.Message.GuildID
	queue := cmd.QueueManager.Get(guildID)
	tracks := queue.List()

	msg := ""

	if queue.CurrentTrack != nil {
		msg += fmt.Sprintf("🎶 Now Playing: %s\n", queue.CurrentTrack.URL)
	} else {
		msg += "📭 Nothing is currently playing.\n"
	}

	if len(tracks) == 0 {
		msg += "🕳️ The queue is empty."
	} else {
		msg += "🎼 Upcoming Queue:\n"
		for i, t := range tracks {
			msg += fmt.Sprintf("%d. %s\n", i+1, t.URL)
		}
	}

	cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, msg)
}

func (cmd *BotCommand) NowPlaying() {
	guildID := cmd.Message.GuildID
	queue := cmd.QueueManager.Get(guildID)

	if queue.CurrentTrack != nil {
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, fmt.Sprintf("🎶 Now Playing: %s", queue.CurrentTrack.URL))
		return
	}

	cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "❌ Nothing is playing.")
}

func (cmd *BotCommand) Pause() {
	guildID := cmd.Message.GuildID
	if state, ok := cmd.AudioSessions.Get(guildID); ok {
		state.Mutex.Lock()
		defer state.Mutex.Unlock()
		if state.Paused {
			cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "⏸️ Already paused.")
			return
		}
		state.Paused = true
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "⏸️ Paused playback.")
	} else {
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "❌ Nothing is playing.")
	}
}

func (cmd *BotCommand) Resume() {
	guildID := cmd.Message.GuildID
	if state, ok := cmd.AudioSessions.Get(guildID); ok {
		state.Mutex.Lock()
		defer state.Mutex.Unlock()
		if !state.Paused {
			cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "▶️ Already playing.")
			return
		}
		state.Paused = false
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "▶️ Resumed playback.")
	} else {
		cmd.Session.ChannelMessageSend(cmd.Message.ChannelID, "❌ Nothing is playing.")
	}
}
