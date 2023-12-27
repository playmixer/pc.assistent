package main

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	jdb "github.com/playmixer/pc.assistent/pkg/jsonstore"
)

var (
	Store *jdb.Driver
)

type CommandType string

const (
	CTExec CommandType = "exec"
)

type StoreCommand struct {
	Commands []string    `json:"commands"`
	Type     CommandType `json:"type"`
	Path     string      `json:"path"`
	Args     []string    `json:"args"`
}

func (command StoreCommand) Init() {
	Store.InitJsonObject(command)
}

func (command StoreCommand) ID() string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(command.Commands, "_")))
}

func (command StoreCommand) Unmarshal(data []byte) (interface{}, error) {
	res := make(map[string]StoreCommand)
	err := json.Unmarshal(data, &res)

	return res, err
}

func StoreInit(name string) (*jdb.Driver, error) {
	var err error
	Store, err = jdb.New(name)
	if err != nil {
		return nil, err
	}

	StoreCommand{}.Init()

	return Store, nil
}
