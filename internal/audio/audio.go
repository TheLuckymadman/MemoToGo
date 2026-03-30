package audio

import (
	"bytes"
	"fmt"
	"os/exec"
)

func OGGToWAV(ogg []byte) ([]byte, error) {
	cmd := exec.Command(
		"ffmpeg",
		"-i", "pipe:0",
		"-ar", "16000",
		"-ac", "1",
		"-f", "wav",
		"pipe:1",
	)

	cmd.Stdin = bytes.NewReader(ogg)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func CnvToOGG(audio []byte) ([]byte, error) {
	cmd := exec.Command(
		"ffmpeg",
		"-i", "pipe:0",
		"-ar", "16000",
		"-ac", "1",
		"-c:a", "libopus",
		"-f", "ogg",
		"pipe:1",
	)

	cmd.Stdin = bytes.NewReader(audio)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %v, %s", err, stderr.String())
	}

	return out.Bytes(), nil
}

func CnvToMP3(audio []byte) ([]byte, error) {
	cmd := exec.Command(
		"ffmpeg",
		"-i", "pipe:0",
		"-ar", "16000", // resample to 16kHz (optional)
		"-ac", "1", // mono (optional)
		"-c:a", "libmp3lame", // MP3 codec
		"-f", "mp3",
		"pipe:1",
	)

	cmd.Stdin = bytes.NewReader(audio)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %v, %s", err, stderr.String())
	}

	return out.Bytes(), nil
}
