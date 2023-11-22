package wavplayer

import (
	"bytes"
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

func PlayWavFromBytes(b []byte) {

	streamer, format, err := wav.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()
	playstream(streamer, format)
}

func PlayWavFromFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := wav.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()
	playstream(streamer, format)
}

func playstream(streamer beep.StreamSeekCloser, format beep.Format) {

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
