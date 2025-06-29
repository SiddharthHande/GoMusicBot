package audio

import "sync"

type QueueManager struct {
	queues map[string]*Queue // guildID → Queue
	sync.RWMutex
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		queues: make(map[string]*Queue),
	}
}

func (qm *QueueManager) Get(guildID string) *Queue {
	qm.Lock()
	defer qm.Unlock()

	if _, ok := qm.queues[guildID]; !ok {
		qm.queues[guildID] = &Queue{}
	}
	return qm.queues[guildID]
}
