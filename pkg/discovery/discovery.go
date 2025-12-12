package discovery

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"slices"

	"github.com/changminbark/golms/pkg/constants"
)

func ListAllLLMs() (map[string][]string, error) {
	// Read ~/golms/ directory
	var golmsPath string
	if homePath, err := os.UserHomeDir(); err != nil {
		return nil, err
	} else {
		golmsPath = path.Join(homePath, "golms")
	}

	// Extract all model server subdirectories
	modelServerList, err := os.ReadDir(golmsPath)
	if err != nil {
		return nil, err
	}

	// Loop through all of the model server subdirectories and add to map
	llmStringMap := make(map[string][]string)
	for _, modelServer := range modelServerList {
		modelServerName := modelServer.Name()
		// If there is an invalid model server directory
		if !slices.Contains(constants.AvailableModelServers, modelServerName) {
			return nil, errors.New("invalid model server directory under ~/golms/")
		}

		// Look at available models under model server directory
		modelServerPath := path.Join(golmsPath, modelServerName)
		llmList, err := os.ReadDir(modelServerPath)
		if err != nil {
			return nil, err
		}

		// Loop through each entry and append to map
		for _, llmEntry := range llmList {
			if llmEntry.IsDir() {
				llmStringMap[modelServerName] = append(llmStringMap[modelServerName], llmEntry.Name())
			}
		}
	}

	return llmStringMap, nil
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
