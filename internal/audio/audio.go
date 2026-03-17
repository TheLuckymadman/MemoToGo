package audio

import (
	"bytes"
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
