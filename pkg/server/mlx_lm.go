package server

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/changminbark/golms/pkg/constants"
	"github.com/changminbark/golms/pkg/ui"
)

type MlxLMServerManager struct {
	BaseModelServerManager
}

func (m *MlxLMServerManager) IsRunning() (bool, int) {
	// Check if python processes are running
	cmd := exec.Command("pgrep", "python")
	pgrepPython, err := cmd.Output()
	if err != nil {
		return false, -1
	}
	pgrepPythonString := string(pgrepPython)
	pgrepPythonStringList := strings.Split(pgrepPythonString, "\n")

	// Check if any of the python processes contain mlx_lm.server
	for _, process := range pgrepPythonStringList {
		process = strings.TrimSpace(process)
		if process == "" {
			continue
		}

		// Use ps to get the full command line for this PID
		cmd = exec.Command("ps", "-p", process, "-o", "command=")
		psOutput, err := cmd.Output()
		if err != nil {
			continue // Process might have terminated
		}

		// Check if the command contains mlx_lm.server
		if strings.Contains(string(psOutput), "mlx_lm.server") {
			pid, _ := strconv.Atoi(process)
			return true, pid
		}
	}
	return false, -1
}

func (m *MlxLMServerManager) Start() error {
	// Build model path
	var modelPath string
	if homePath, err := os.UserHomeDir(); err != nil {
		return err
	} else {
		modelPath = path.Join(homePath, "/golms", constants.Mlx_lm, m.llm)
	}

	// Check if model path exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model path does not exist: %s", modelPath)
	}

	// Create log file for server output
	logFile, err := os.Create("/tmp/mlx_lm_server.log")
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	// Set the port we'll use
	m.port = 8080

	// Run command in background
	cmd := exec.Command("mlx_lm.server", "--model", modelPath, "--host", "127.0.0.1", "--port", strconv.Itoa(m.port))
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start mlx_lm.server: %w", err)
	}

	fmt.Println(ui.SubtleStyle.Render(fmt.Sprintf("Server started with PID: %d", cmd.Process.Pid)))
	fmt.Println(ui.SubtleStyle.Render("Logs: /tmp/mlx_lm_server.log"))

	// Wait for server to start listening on the port
	fmt.Print(ui.SubtleStyle.Render("Waiting for server to initialize"))
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)

		// Check if server is listening on the port
		running, pid := m.IsRunning()
		if running {
			// Verify port is listening
			checkCmd := exec.Command("lsof", "-Pan", "-p", strconv.Itoa(pid), "-i")
			if output, err := checkCmd.Output(); err == nil && len(output) > 0 {
				// Port is listening
				fmt.Println()
				fmt.Println(ui.FormatSuccess(fmt.Sprintf("Server is listening on port %d", m.port)))
				return nil
			}
		}
	}
	fmt.Println()
	fmt.Println(ui.FormatWarning("Server process started but may not be listening yet"))
	fmt.Println(ui.SubtleStyle.Render("Check logs at /tmp/mlx_lm_server.log"))

	return nil
}

func (m *MlxLMServerManager) Stop() error {
	// Check if running
	isRunning, pid := m.IsRunning()
	if !isRunning {
		return errors.New("mlx_lm.server is not running")
	}
	// Kill PID
	if err := exec.Command("kill", "-9", strconv.Itoa(pid)).Run(); err != nil {
		return fmt.Errorf("failed to kill process with pid %d: %w", pid, err)
	}
	// Reset port
	m.port = 0
	return nil
}

func (m *MlxLMServerManager) GetPort() (int, error) {
	return getPortHelper(m, m.port)
}
