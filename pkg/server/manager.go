package server

import "github.com/changminbark/golms/pkg/constants"

type ModelServerManager interface {
	IsAvailable() bool
	IsRunning() (bool, int)
	Start() error
	Stop() error
	GetPort() (int, error)
}

func NewServerManager(model_server string, llm string) ModelServerManager {
	switch model_server {
	case constants.Mlx_lm:
		return &MlxLMServerManager{llm: llm, port: 0}
	default:
		return nil
	}
}
