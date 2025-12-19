package server

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strconv"

	"github.com/changminbark/golms/pkg/constants"
	"github.com/changminbark/golms/pkg/discovery"
)

type ModelServerManager interface {
	IsAvailable() bool
	IsRunning() (bool, int)
	Start() error
	Stop() error
	GetPort() (int, error)
}

type BaseModelServerManager struct {
	modelServer string
	llm         string
	port        int
}

func (m *BaseModelServerManager) IsAvailable() bool {
	modelServerList, _ := discovery.ListAllModelServers()
	return slices.Contains(modelServerList, m.modelServer)
}

func NewServerManager(model_server string, llm string) ModelServerManager {
	switch model_server {
	case constants.Mlx_lm:
		return &MlxLMServerManager{
			BaseModelServerManager: BaseModelServerManager{
				modelServer: constants.Mlx_lm,
				llm:         llm,
				port:        0,
			},
		}
	case constants.Ollama:
		return nil
	default:
		return nil
	}
}

// getPortHelper is a helper function that can be used by specific implementations
// to get the port for a running server
func getPortHelper(mgr ModelServerManager, knownPort int) (int, error) {
	// Check if running
	isRunning, pid := mgr.IsRunning()
	if !isRunning {
		return -1, errors.New("Model Server is not running")
	}

	// If we started the server and know the port, return it
	if knownPort > 0 {
		return knownPort, nil
	}

	// Otherwise, try to detect the port using lsof (for externally started servers)
	cmd := exec.Command("lsof", "-Pan", "-p", strconv.Itoa(pid), "-i")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return -1, fmt.Errorf("failed to get port info: %w (try running: lsof -Pan -p %d -i)", err, pid)
	}

	// Parse port number from output
	re := regexp.MustCompile(`:(\d+)\s+\(LISTEN\)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return -1, errors.New("no listening port found in lsof output")
	}
	port, err := strconv.Atoi(matches[1]) // This is the first group in regex match (\d+)
	if err != nil {
		return -1, fmt.Errorf("failed to parse port number: %w", err)
	}

	return port, nil
}

