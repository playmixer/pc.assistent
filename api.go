package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/playmixer/corvid/listen"
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

	_cmd, err := Store.Open(Command{}).All()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	commands, ok := _cmd.(map[string]Command)
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

	data := Command{
		Commands: payload.Commands,
		Type:     CommandType(payload.Type),
		Path:     payload.Path,
		Args:     payload.Args,
	}
	err := Store.Open(Command{}).Insert(data)
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

	err := Store.Open(Command{}).Remove(Command{
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

	_res, err := Store.Open(Command{}).All()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	res := []commandListGetResponse{}
	for _, v := range _res.(map[string]Command) {
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

	data := Command{
		Commands: payload.Commands,
		Type:     CommandType(payload.Type),
		Path:     payload.Path,
		Args:     payload.Args,
	}
	_res, err := Store.Open(Command{}).Update(payload.ID, data)
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
	_mapRes := make(map[string]Command)
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

func httpInfo(ctx *gin.Context) {

	result := map[string]interface{}{
		"names":    assistent.Names,
		"commands": assistent.GetCommands(),
		"deviceId": assistent.GetRecorder().DeviceId,
	}

	result["devices"], _ = listen.GetMicrophons()

	ctx.JSON(200, result)
}

func httpRefresh(ctx *gin.Context) {
	body := map[string][]string{}
	err := ctx.ShouldBindJSON(&body)
	if err != nil {
		ctx.JSON(500, response{
			Status: false,
			Error:  err.Error(),
		})
		return
	}

	refresh, ok := body["request"]
	if !ok {
		ctx.JSON(204, response{
			Status:  true,
			Message: "not found attr request",
		})
		return
	}
	for _, v := range refresh {
		if v == "commands" {
			err = LoadCommand()
		}

		if err != nil {
			ctx.JSON(500, response{
				Status: false,
				Error:  err.Error(),
			})
			return
		}
	}

	ctx.JSON(200, response{
		Status: true,
	})
}
