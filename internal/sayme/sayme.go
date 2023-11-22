package sayme

import (
	"pc.assistent/internal/wavplayer"
)

const (
	MediaListen     = "./media/listen2.wav"
	MediaHello      = "./media/hello.wav"
	MediaSpaceScrim = "./media/kosmicheskiy-skrip.wav"
)

type SayMe struct{}

func New() *SayMe {
	return &SayMe{}
}

func (s *SayMe) TalkListen() {
	wavplayer.PlayWavFromFile(MediaListen)
}

func (s *SayMe) TalkHello() {
	wavplayer.PlayWavFromFile(MediaHello)
}

func (s *SayMe) SoundStart() {
	wavplayer.PlayWavFromFile(MediaSpaceScrim)
}

func (s *SayMe) SoundEnd() {
	wavplayer.PlayWavFromFile(MediaSpaceScrim)
}
