package audio

import (
	"os/exec"
	"strings"
	"sync"
)

type LoopMode int

const (
	LoopOff LoopMode = iota
	LoopOne
	LoopAll
)

func (m LoopMode) String() string {
	switch m {
	case LoopOff:
		return "off"
	case LoopOne:
		return "one"
	case LoopAll:
		return "all"
	default:
		return "unknown"
	}
}

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
	LoopMode     LoopMode
}

func (q *Queue) Enqueue(t *Track) {
	q.Lock()
	defer q.Unlock()
	q.Tracks = append(q.Tracks, t)
}

func (q *Queue) EnqueueMultiple(tracks []*Track) {
	q.Lock()
	defer q.Unlock()
	q.Tracks = append(q.Tracks, tracks...)
}

func (q *Queue) Dequeue() *Track {
	q.Lock()
	defer q.Unlock()

	if q.CurrentTrack != nil && q.LoopMode == LoopOne {
		return q.CurrentTrack
	}

	if len(q.Tracks) == 0 {
		return nil
	}

	track := q.Tracks[0]
	q.Tracks = q.Tracks[1:]

	if q.LoopMode == LoopAll {
		q.Tracks = append(q.Tracks, track)
	}

	q.CurrentTrack = track
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
	q.CurrentTrack = nil
}

func (q *Queue) SetLoopMode(mode LoopMode) {
	q.Lock()
	defer q.Unlock()
	q.LoopMode = mode
}

func (q *Queue) GetLoopMode() LoopMode {
	q.Lock()
	defer q.Unlock()
	return q.LoopMode
}

func (q *Queue) ToggleLoopMode() LoopMode {
	q.Lock()
	defer q.Unlock()
	q.LoopMode = (q.LoopMode + 1) % 3
	return q.LoopMode
}

func ExtractPlaylistTracks(playlistURL string) ([]*Track, error) {
	cmd := exec.Command("yt-dlp", "--flat-playlist", "--print", "%(title)s|%(url)s", playlistURL)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var tracks []*Track
	for _, line := range lines {
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		title, url := parts[0], parts[1]
		// Fix: Only prepend if not already a full URL
		if !strings.HasPrefix(url, "http") {
			url = "https://www.youtube.com/watch?v=" + url
		}
		tracks = append(tracks, &Track{
			Title: title,
			URL:   url,
		})
	}
	return tracks, nil
}
