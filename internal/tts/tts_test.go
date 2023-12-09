package tts

import "testing"

func TestVoice(t *testing.T) {
	speach := New()
	speach.SetPath("..\\..\\voice.bat")
	if err := speach.Voice("привет"); err != nil {
		t.Fatalf(err.Error())
	}
}
