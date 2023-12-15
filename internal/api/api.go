package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/playmixer/pc.assistent/internal/store"
)

func WSHandle(message []byte) {
	fmt.Println(string(message))
}

type response struct {
	Status  bool        `json:"status"`
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type commandListGetResponse struct {
	ID       string   `json:"id"`
	Commands []string `json:"commands"`
	Type     string   `json:"type"`
	Path     string   `json:"path"`
	Args     []string `json:"args"`
}

func GetCommands(c *gin.Context) {

	db := store.Store
	_cmd, err := db.Open(store.Command{}).All()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	commands, ok := _cmd.(map[string]store.Command)
	if !ok {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  "Conversion failed",
		})
		return
	}
	var result []commandListGetResponse
	for k, c := range commands {
		_r := commandListGetResponse{
			ID:       k,
			Commands: c.Commands,
			Type:     string(c.Type),
			Path:     c.Path,
			Args:     c.Args,
		}
		result = append(result, _r)
	}

	c.JSON(http.StatusOK, result)
}

func NewCommand(c *gin.Context) {
	payload := commandListGetResponse{}
	c.ShouldBindJSON(&payload)

	if len(payload.Commands) == 0 {
		c.JSON(http.StatusBadRequest, response{
			Status: false,
			Error:  "commands not found",
		})
		return
	}

	db := store.Store
	data := store.Command{
		Commands: payload.Commands,
		Type:     store.CommandType(payload.Type),
		Path:     payload.Path,
		Args:     payload.Args,
	}
	err := db.Open(store.Command{}).Insert(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	res := commandListGetResponse{
		ID:       data.ID(),
		Commands: data.Commands,
		Type:     string(data.Type),
		Path:     data.Path,
		Args:     data.Args,
	}

	c.JSON(http.StatusOK, response{
		Status: true,
		Data:   res,
	})
}

func DeleteCommand(c *gin.Context) {
	payload := commandListGetResponse{}
	c.ShouldBindJSON(&payload)

	if payload.ID == "" {
		c.JSON(http.StatusBadRequest, response{
			Status: false,
			Error:  "id requierd",
		})
		return
	}

	db := store.Store
	err := db.Open(store.Command{}).Remove(store.Command{
		Commands: payload.Commands,
		Type:     store.CommandType(payload.Type),
		Path:     payload.Path,
		Args:     payload.Args,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	_res, err := db.Open(store.Command{}).All()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	res := []commandListGetResponse{}
	for _, v := range _res.(map[string]store.Command) {
		res = append(res, commandListGetResponse{
			ID:       v.ID(),
			Commands: v.Commands,
			Type:     string(v.Type),
			Path:     v.Path,
			Args:     v.Args,
		})
	}

	c.JSON(http.StatusOK, response{
		Status: true,
		Data:   res,
	})
}

func UpdateCommand(c *gin.Context) {
	payload := commandListGetResponse{}
	c.ShouldBindJSON(&payload)

	if payload.ID == "" {
		c.JSON(http.StatusBadRequest, response{
			Status: false,
			Error:  "id requierd",
		})
		return
	}

	if len(payload.Commands) == 0 {
		c.JSON(http.StatusBadRequest, response{
			Status: false,
			Error:  "commands not found",
		})
		return
	}

	db := store.Store
	data := store.Command{
		Commands: payload.Commands,
		Type:     store.CommandType(payload.Type),
		Path:     payload.Path,
		Args:     payload.Args,
	}
	_res, err := db.Open(store.Command{}).Update(payload.ID, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	_b, err := json.Marshal(_res)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}
	_mapRes := make(map[string]store.Command)
	err = json.Unmarshal(_b, &_mapRes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	res := []commandListGetResponse{
		commandListGetResponse{
			ID:       data.ID(),
			Commands: data.Commands,
			Type:     string(data.Type),
			Path:     data.Path,
			Args:     data.Args,
		},
	}

	c.JSON(http.StatusOK, response{
		Status: true,
		Data:   res,
	})
}
