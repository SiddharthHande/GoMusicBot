package audio

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

const (
	CHANNELS   = 2
	FRAME_RATE = 48000
	FRAME_SIZE = 960
	MAX_BYTES  = (FRAME_SIZE * 2) * 2
)

type Connection struct {
	voiceConnection *discordgo.VoiceConnection
	send            chan []int16
	lock            sync.Mutex
	sendpcm         bool
	stopRunning     bool
	playing         bool
	ffmpegCmd       *exec.Cmd
	ytdlpCmd        *exec.Cmd
}

func NewConnection(voiceConnection *discordgo.VoiceConnection) *Connection {
	return &Connection{
		voiceConnection: voiceConnection,
	}
}

func (connection *Connection) Disconnect() {
	connection.voiceConnection.Disconnect()
}

func (connection *Connection) sendPCM(voice *discordgo.VoiceConnection, pcm <-chan []int16, paused *bool, pauseMutex *sync.Mutex) {
	connection.lock.Lock()
	if connection.sendpcm || pcm == nil {
		connection.lock.Unlock()
		return
	}
	connection.sendpcm = true
	connection.lock.Unlock()
	defer func() {
		connection.lock.Lock()
		connection.sendpcm = false
		connection.lock.Unlock()
		if r := recover(); r != nil {
			fmt.Println("Recovered in sendPCM:", r)
		}
	}()

	encoder, err := gopus.NewEncoder(FRAME_RATE, CHANNELS, gopus.Audio)
	if err != nil {
		fmt.Println("NewEncoder error,", err)
		return
	}

	for frame := range pcm {
		// Pause logic
		for {
			pauseMutex.Lock()
			isPaused := *paused
			pauseMutex.Unlock()
			if !isPaused {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		opus, err := encoder.Encode(frame, FRAME_SIZE, MAX_BYTES)
		if err != nil {
			fmt.Println("Encoding error,", err)
			return
		}
		if !voice.Ready || voice.OpusSend == nil {
			fmt.Printf("Discordgo not ready for opus packets. %+v : %+v", voice.Ready, voice.OpusSend)
			return
		}
		voice.OpusSend <- opus
	}

	fmt.Println("sendPCM: channel closed, exiting")
}

func (connection *Connection) Play(youtubeURL string, paused *bool, pauseMutex *sync.Mutex) error {
	connection.lock.Lock()
	if connection.playing {
		connection.lock.Unlock()
		return errors.New("song already playing")
	}
	connection.playing = true
	connection.stopRunning = false
	connection.lock.Unlock()

	ytdlp := exec.Command("yt-dlp", "-f", "bestaudio", "-o", "-", youtubeURL)
	ffmpeg := exec.Command("ffmpeg",
		"-re",
		"-i", "pipe:0",
		"-f", "s16le",
		"-ar", strconv.Itoa(FRAME_RATE),
		"-ac", strconv.Itoa(CHANNELS),
		"pipe:1",
	)

	connection.ffmpegCmd = ffmpeg
	connection.ytdlpCmd = ytdlp
	ytdlp.Stderr = os.Stderr
	ffmpeg.Stderr = os.Stderr

	ytdlpOut, err := ytdlp.StdoutPipe()
	if err != nil {
		fmt.Println("yt-dlp pipe error:", err)
		return err
	}
	ffmpeg.Stdin = ytdlpOut

	out, err := ffmpeg.StdoutPipe()
	if err != nil {
		return err
	}
	buffer := bufio.NewReaderSize(out, 16384)

	if err := ytdlp.Start(); err != nil {
		fmt.Println("yt-dlp start error:", err)
		return err
	}
	if err := ffmpeg.Start(); err != nil {
		return err
	}

	connection.voiceConnection.Speaking(true)
	defer func() {
		connection.voiceConnection.Speaking(false)
		connection.lock.Lock()
		connection.playing = false
		connection.lock.Unlock()
	}()

	connection.lock.Lock()
	if connection.send != nil {
		close(connection.send)
	}
	connection.send = make(chan []int16, 2)
	sendChan := connection.send
	connection.lock.Unlock()

	go connection.sendPCM(connection.voiceConnection, sendChan, paused, pauseMutex)

	for {
		connection.lock.Lock()
		if connection.stopRunning {
			_ = ffmpeg.Process.Kill()
			connection.lock.Unlock()
			break
		}
		send := connection.send
		connection.lock.Unlock()

		audioBuffer := make([]int16, FRAME_SIZE*CHANNELS)
		err = binary.Read(buffer, binary.LittleEndian, &audioBuffer)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}
		if send != nil {
			select {
			case send <- audioBuffer:
			default:
				// drop frame
			}
		}
	}

	return nil
}

func (connection *Connection) Stop() {
	connection.lock.Lock()
	if connection.stopRunning {
		connection.lock.Unlock()
		return
	}
	connection.stopRunning = true
	connection.playing = false
	if connection.ffmpegCmd != nil && connection.ffmpegCmd.Process != nil {
		_ = connection.ffmpegCmd.Process.Kill()
	}
	if connection.ytdlpCmd != nil && connection.ytdlpCmd.Process != nil {
		_ = connection.ytdlpCmd.Process.Kill()
	}
	send := connection.send
	connection.send = nil
	connection.lock.Unlock()

	if send != nil {
		close(send)
	}
}
