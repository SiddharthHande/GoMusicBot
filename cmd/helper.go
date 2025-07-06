package commands

import (
	"musicbot/audio"
	"os/exec"
	"strings"
)

func extractMetadata(t *audio.Track) {
	cmdYTDLP := exec.Command("yt-dlp", "--quiet", "--no-warnings", "--no-playlist",
		"--print", "%(title)s|%(duration_string)s|%(uploader)s", t.URL)
	if output, err := cmdYTDLP.Output(); err == nil {
		parts := strings.SplitN(strings.TrimSpace(string(output)), "|", 3)
		if len(parts) == 3 {
			t.Title = parts[0]
			t.Duration = parts[1]
			t.Uploader = parts[2]
		}
	}
}
