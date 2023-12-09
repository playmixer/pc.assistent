package tts

import (
	"os/exec"
	"strings"
)

type TTS struct {
	f string
}

func New() *TTS {

	tts := TTS{
		f: ".\\voice.bat",
	}

	return &tts
}

func (t *TTS) Voice(s string) error {
	sArr := strings.Fields(s)
	for _, word := range sArr {
		if err := exec.Command(t.f, word).Run(); err != nil {
			return err
		}
	}
	return nil
}

func (t *TTS) SetPath(p string) {
	t.f = p
}
