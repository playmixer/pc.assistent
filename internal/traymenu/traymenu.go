package traymenu

import (
	"log"
	"os"

	"github.com/getlantern/systray"
)

const (
	iconDefaultPath = ".\\default.ico"
)

var (
	iconDefaultByte = []byte{}
)

func Init() {
	var err error
	iconDefaultByte, err = os.ReadFile(iconDefaultPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	go systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconDefaultByte)
	systray.SetTitle("PC.Assistent App")
	systray.SetTooltip("A App")
}

func onExit() {
	systray.Quit()
}

func SetIconDefault() {
	systray.SetIcon(iconDefaultByte)
}

func SetIcon(path string) {
	iconByte, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err.Error())
	}
	systray.SetIcon(iconByte)
}
