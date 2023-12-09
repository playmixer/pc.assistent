package smarthome

import "errors"

type TypeSHServiceName string

const (
	SHTuyaService TypeSHServiceName = "tuya"
)

type GetDeviceResponse struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
	T       int64       `json:"t"`
}

type PostDeviceCmdResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
	Result  bool   `json:"result"`
	T       int64  `json:"t"`
}

type IApiCommand interface {
	Add(code string, value bool) IApiCommand
}

type IApi interface {
	GetDevice(device string) GetDeviceResponse
	PostDevice(deviceId string, commands IApiCommand)
	NewCommand() IApiCommand
}

func FactoryNew(name TypeSHServiceName) (IApi, error) {
	if name == SHTuyaService {
		return NewTuyaApi(), nil
	}

	return nil, errors.New("Factory not found service: " + string(name))
}
