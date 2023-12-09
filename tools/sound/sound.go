package main

import (
	"fmt"
	"log"
	"os"

	v "github.com/itchyny/volume-go"
)

func GetVolume() int {
	vol, err := v.GetVolume()
	if err != nil {
		log.Fatalf("get volume failed: %+v", err)
	}
	fmt.Printf("current volume: %v\n", vol)
	return vol
}

func SetVolume(vol int) {
	err := v.SetVolume(vol)
	if err != nil {
		log.Fatalf("set volume failed: %+v", err)
	}
	fmt.Printf("set volume %v success\n", vol)
}

func main() {

	args := os.Args
	if len(args) < 2 {
		return
	}
	v := args[1]
	vol := GetVolume()
	if v == "up" {
		SetVolume(vol + 10)
	}
	if v == "down" {
		SetVolume(vol - 10)
	}
}
