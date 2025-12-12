package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/changminbark/golms/pkg/client"
	"github.com/changminbark/golms/pkg/constants"
	"github.com/changminbark/golms/pkg/discovery"
	"github.com/changminbark/golms/pkg/server"
)

func NewCLI() *cobra.Command {
	// Create root command where user types golms
	rootCmd := &cobra.Command{
		Use:           "golms",
		Short:         "Local Model Server Interface written in Go",
		SilenceUsage:  true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print(cmd.UsageString())
		},
	}

	// Create list command that lists available LLMs and model servers
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available LLMs and model servers",
		RunE:  listHandler,
	}

	// Create servers command that lists support model servers
	serversCmd := &cobra.Command{
		Use:   "servers",
		Short: "List all supported model servers",
		Run:   serversHandler,
	}

	// Create connect command that will connect to model server and LLM
	connectCmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to a model server and LLM",
		RunE:  connectHandler,
	}

	// Add subcommands to root command
	rootCmd.AddCommand(listCmd, serversCmd, connectCmd)

	return rootCmd
}

// ==================== Command Handlers ====================
func listHandler(cmd *cobra.Command, args []string) error {
	// Get list of all LLMs
	llmListMap, err := discovery.ListAllLLMs()
	if err != nil {
		fmt.Printf("Error encountered while listing LLMS: %v", err)
		return err
	}
	if len(llmListMap) == 0 {
		fmt.Print("You have no model server directories available in the ~/golms directory. \nMake sure to download them as folders with model weights.")
		return errors.New("no model server directories found")
	}

	// Print available LLMs
	fmt.Print("You have the following LLMs available:\n")
	for modelServer, llmList := range llmListMap {
		fmt.Printf("- %s:\n", modelServer)
		for _, llm := range llmList {
			fmt.Printf("  - %s\n", llm)
		}
	}

	// Spacing
	fmt.Println()

	// Get list of all model servers
	modelServerList, err := discovery.ListAllModelServers()
	if err != nil {
		fmt.Printf("Error encountered while listing model servers: %v\n", err)
		return err
	}
	if len(modelServerList) == 0 {
		fmt.Print("You have no model servers available. \nMake sure to download the following supported models:\n")
		for _, modelServer := range constants.AvailableModelServers {
			fmt.Printf("- %s\n", modelServer)
		}
		return errors.New("no model servers found")
	}

	// Print available model servers
	fmt.Print("You have the following model servers available:\n")
	for _, modelServer := range modelServerList {
		fmt.Printf("- %s\n", modelServer)
	}

	return nil
}

func serversHandler(cmd *cobra.Command, args []string) {
	fmt.Print("The following model servers are supported:\n")
	for _, modelServer := range constants.AvailableModelServers {
		fmt.Printf("- %s\n", modelServer)
	}
}

func connectHandler(cmd *cobra.Command, args []string) error {
	// Initialize data objects
	reader := bufio.NewReader(os.Stdin)
	modelServerMap := make(map[int]string)
	llmMap := make(map[int]string)
	var selectedModelServer string
	var selectedLLM string

	// Get list of all model servers
	modelServerList, err := discovery.ListAllModelServers()
	if err != nil {
		fmt.Printf("Error encountered while listing model servers: %v\n", err)
		return err
	}
	if len(modelServerList) == 0 {
		fmt.Print("You have no model servers available. \nMake sure to download the following supported models:\n")
		for _, modelServer := range constants.AvailableModelServers {
			fmt.Printf("- %s\n", modelServer)
		}
		return errors.New("no model servers found")
	}

	// Let user choose a model server
	fmt.Print("Choose one of the model servers you have available (type just the number):\n")
	for idx, modelServer := range modelServerList {
		modelServerMap[idx] = modelServer
		fmt.Printf("%d. %s\n", idx, modelServer)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	inputNum, err := strconv.Atoi(input)
	if err != nil {
		fmt.Print("Please input just the number corresponding to the model server and press enter\n")
		return err
	}
	selectedModelServer, ok := modelServerMap[inputNum]
	if !ok {
		fmt.Printf("The following input number is not valid: %d\n", inputNum)
		return errors.New("invalid input number")
	}

	// Get following LLMs for that model server
	llmListMap, err := discovery.ListAllLLMs()
	if err != nil {
		fmt.Printf("Error encountered while listing LLMS: %v\n", err)
		return err
	}
	if len(llmListMap) == 0 {
		fmt.Print("You have no model server directories available in the ~/golms directory. \nMake sure to download them as folders with model weights.")
		return errors.New("no model server directories found")
	}
	llmList, ok := llmListMap[selectedModelServer]
	if !ok {
		fmt.Printf("You do not have a subdirectory under ~/golms for the model server: %s\n", selectedModelServer)
		return fmt.Errorf("no subdirectory found for model server: %s", selectedModelServer)
	}
	if len(llmList) == 0 {
		fmt.Printf("You do not have any LLMs available for the following model server: %s\n", selectedModelServer)
		return fmt.Errorf("no llms available for model server: %s", selectedModelServer)
	}

	// Let user choose LLM
	fmt.Print("You have the following LLMs available. Please input just the number corresponding to the LLM and press enter:\n")
	for idx, llm := range llmList {
		llmMap[idx] = llm
		fmt.Printf("%d. %s\n", idx, llm)
	}
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	inputNum, err = strconv.Atoi(input)
	if err != nil {
		fmt.Print("Please input just the number corresponding to the LLM and press enter\n")
		return err
	}
	selectedLLM, ok = llmMap[inputNum]
	if !ok {
		fmt.Printf("The following input number is not valid: %d\n", inputNum)
		return errors.New("invalid input number")
	}

	// Create model server instance
	modelServerManager := server.NewServerManager(selectedModelServer, selectedLLM)
	if !modelServerManager.IsAvailable() {
		fmt.Printf("The following model server is not available: %s\n", selectedModelServer)
		return errors.New("model server unavailable")
	}
	running, _ := modelServerManager.IsRunning()
	if !running {
		err := modelServerManager.Start()
		if err != nil {
			return err
		}
		defer modelServerManager.Stop()
	}
	port, err := modelServerManager.GetPort()
	if err != nil {
		fmt.Print("Failed to get port in ModelServerManager\n")
		return err
	}

	// Create client to communicate with model server
	modelServerClient := client.NewClient(selectedModelServer, selectedLLM, constants.Localhost, port, reader)

	// Start chat
	err = modelServerClient.StartChat()
	if err != nil {
		return err
	}

	// Will clean up with defer of modelServerManager.Stop()
	return nil
}
