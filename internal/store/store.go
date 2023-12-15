package store

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

type Command struct {
	Commands []string    `json:"commands"`
	Type     CommandType `json:"type"`
	Path     string      `json:"path"`
	Args     []string    `json:"args"`
}

func (command Command) Init() {
	Store.InitJsonObject(command)
}

func (command Command) ID() string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(command.Commands, "_")))
}

func (command Command) Unmarshal(data []byte) (interface{}, error) {
	res := make(map[string]Command)
	err := json.Unmarshal(data, &res)

	return res, err
}

func Init(name string) (*jdb.Driver, error) {
	var err error
	Store, err = jdb.New(name)
	if err != nil {
		return nil, err
	}

	Command{}.Init()

	return Store, nil
}
