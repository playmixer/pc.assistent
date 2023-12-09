package main

import (
	"log"

	"github.com/zserge/lorca"
)

func main() {
	// curDir, err := os.Getwd()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	args := []string{}
	args = append(args, "--remote-allow-origins=*")
	url := "http://localhost:3000/"
	ui, err := lorca.New(url, "",
		400, 700, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	<-ui.Done()
}
