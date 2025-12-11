package discovery

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/changminbark/golms/pkg/constants"
)

func ListAllLLMs() ([]string, error) {
	// Read ~/LLMs/ directory
	var llmPath string
	if homePath, err := os.UserHomeDir(); err != nil {
		return nil, err
	} else {
		llmPath = path.Join(homePath, "LLMs")
	}

	llmList, err := os.ReadDir(llmPath)
	if err != nil {
		return nil, err
	}

	var llmStringList []string
	for _, llmEntry := range llmList {
		if llmEntry.IsDir() {
			llmStringList = append(llmStringList, llmEntry.Name())
		}
	}

	return llmStringList, err
}

func ListAllModelServers() ([]string, error) {
	var modelServerStringList []string
	// Check if ollama exists
	if isBinaryAvailable(constants.Ollama) {
		modelServerStringList = append(modelServerStringList, constants.Ollama)
	}
	// Check if mlx_lm exists
	if isPythonModuleAvailable(constants.Mlx_lm) {
		modelServerStringList = append(modelServerStringList, constants.Mlx_lm)
	}

	return modelServerStringList, nil
}

func isBinaryAvailable(name string) bool {
	rootBinPath := "/bin/" + name
	localUsrBinPath := "/usr/local/bin/" + name
	_, errRoot := exec.LookPath(rootBinPath)
	_, errUsr := exec.LookPath(localUsrBinPath)

	return errRoot == nil || errUsr == nil
}

func isPythonModuleAvailable(module string) bool {
	cmd := exec.Command("python3", "-c", fmt.Sprintf("import %s", module))
	return cmd.Run() == nil
}