package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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

func httpGetCommands(c *gin.Context) {

	_cmd, err := Store.Open(StoreCommand{}).All()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	commands, ok := _cmd.(map[string]StoreCommand)
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

func httpNewCommand(c *gin.Context) {
	payload := commandListGetResponse{}
	c.ShouldBindJSON(&payload)

	if len(payload.Commands) == 0 {
		c.JSON(http.StatusBadRequest, response{
			Status: false,
			Error:  "commands not found",
		})
		return
	}

	data := StoreCommand{
		Commands: payload.Commands,
		Type:     CommandType(payload.Type),
		Path:     payload.Path,
		Args:     payload.Args,
	}
	err := Store.Open(StoreCommand{}).Insert(data)
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

func httpDeleteCommand(c *gin.Context) {
	payload := commandListGetResponse{}
	c.ShouldBindJSON(&payload)

	if payload.ID == "" {
		c.JSON(http.StatusBadRequest, response{
			Status: false,
			Error:  "id requierd",
		})
		return
	}

	err := Store.Open(StoreCommand{}).Remove(StoreCommand{
		Commands: payload.Commands,
		Type:     CommandType(payload.Type),
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

	_res, err := Store.Open(StoreCommand{}).All()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	res := []commandListGetResponse{}
	for _, v := range _res.(map[string]StoreCommand) {
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

func httpUpdateCommand(c *gin.Context) {
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

	data := StoreCommand{
		Commands: payload.Commands,
		Type:     CommandType(payload.Type),
		Path:     payload.Path,
		Args:     payload.Args,
	}
	_res, err := Store.Open(StoreCommand{}).Update(payload.ID, data)
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
	_mapRes := make(map[string]StoreCommand)
	err = json.Unmarshal(_b, &_mapRes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	res := []commandListGetResponse{
		{
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
