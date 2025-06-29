package vc

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

type VoiceManager struct {
	mu          sync.RWMutex
	connections map[string]*discordgo.VoiceConnection // guildID → VC
}

func NewVoiceManager() *VoiceManager {
	return &VoiceManager{
		connections: make(map[string]*discordgo.VoiceConnection),
	}
}

// Join joins the voice channel and stores the connection.
func (vm *VoiceManager) Join(s *discordgo.Session, guildID, channelID string) (*discordgo.VoiceConnection, error) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, err
	}

	vm.mu.Lock()
	vm.connections[guildID] = vc
	vm.mu.Unlock()

	return vc, nil
}

// Leave disconnects and removes the VC.
func (vm *VoiceManager) Leave(guildID string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vc, ok := vm.connections[guildID]
	if !ok {
		return nil // nothing to disconnect
	}

	err := vc.Disconnect()
	delete(vm.connections, guildID)

	return err
}

// Get returns the VC for the guild, if it exists.
func (vm *VoiceManager) Get(guildID string) (*discordgo.VoiceConnection, bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	vc, ok := vm.connections[guildID]
	return vc, ok
}
