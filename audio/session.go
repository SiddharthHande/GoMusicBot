package audio

import "sync"

type GuildAudioState struct {
	Conn   *Connection
	Mutex  sync.Mutex
	Paused bool
}

type AudioSessionManager struct {
	sessions map[string]*GuildAudioState
	mutex    sync.RWMutex
}

func NewAudioSessionManager() *AudioSessionManager {
	return &AudioSessionManager{
		sessions: make(map[string]*GuildAudioState),
	}
}

func (m *AudioSessionManager) Get(guildID string) (*GuildAudioState, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	state, ok := m.sessions[guildID]
	return state, ok
}

func (m *AudioSessionManager) Set(guildID string, state *GuildAudioState) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.sessions[guildID] = state
}

func (m *AudioSessionManager) Delete(guildID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.sessions, guildID)
}
