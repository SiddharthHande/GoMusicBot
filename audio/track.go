package audio

import (
	"sync"
)

type Track struct {
	URL      string
	Title    string
	Duration string
	Uploader string
}

type Queue struct {
	Tracks []*Track
	sync.Mutex
	IsPlaying    bool
	CurrentTrack *Track
	LoopTrack    bool
	LoopQueue    bool
}

func (q *Queue) Enqueue(t *Track) {
	q.Lock()
	defer q.Unlock()
	q.Tracks = append(q.Tracks, t)
}

func (q *Queue) Dequeue() *Track {
	q.Lock()
	defer q.Unlock()

	if len(q.Tracks) == 0 {
		return nil
	}
	track := q.Tracks[0]
	q.Tracks = q.Tracks[1:]
	return track
}

func (q *Queue) List() []*Track {
	q.Lock()
	defer q.Unlock()
	return append([]*Track(nil), q.Tracks...)
}

func (q *Queue) Clear() {
	q.Lock()
	defer q.Unlock()
	q.Tracks = nil
}

func (q *Queue) Play() {
	for {
		track := q.Dequeue()
		if track == nil {
			break
		}
		q.CurrentTrack = track
		// ...play...
		if q.LoopTrack {
			q.Lock()
			q.Tracks = append([]*Track{track}, q.Tracks...)
			q.Unlock()
		}
		if q.LoopQueue {
			q.Lock()
			q.Tracks = append(q.Tracks, track)
			q.Unlock()
		}
	}
}
