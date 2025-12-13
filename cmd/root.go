package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/changminbark/golms/pkg/client"
	"github.com/changminbark/golms/pkg/constants"
	"github.com/changminbark/golms/pkg/discovery"
	"github.com/changminbark/golms/pkg/server"
	"github.com/changminbark/golms/pkg/ui"
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
		fmt.Println(ui.FormatError(fmt.Sprintf("Error encountered while listing LLMs: %v", err)))
		return err
	}
	if len(llmListMap) == 0 {
		fmt.Println(ui.FormatWarning("No model server directories found"))
		fmt.Println(ui.SubtleStyle.Render("Make sure models are placed in ~/golms/<model_server>/ directories"))
		return errors.New("no model server directories found")
	}

	// Print available LLMs
	fmt.Println(ui.FormatHeader("Available LLMs", "Models organized by server"))
	for modelServer, llmList := range llmListMap {
		fmt.Println(ui.FormatListItem(modelServer + ":"))
		for _, llm := range llmList {
			fmt.Println(ui.FormatNestedListItem(llm))
		}
	}

	// Spacing
	fmt.Println()

	// Get list of all model servers
	modelServerList, err := discovery.ListAllModelServers()
	if err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error encountered while listing model servers: %v", err)))
		return err
	}
	if len(modelServerList) == 0 {
		fmt.Println(ui.FormatWarning("No model servers available"))
		fmt.Println(ui.SubtleStyle.Render("Install one of the following supported model servers:"))
		for _, modelServer := range constants.AvailableModelServers {
			fmt.Println(ui.FormatListItem(modelServer))
		}
		return errors.New("no model servers found")
	}

	// Print available model servers
	fmt.Println(ui.FormatHeader("Installed Model Servers"))
	for _, modelServer := range modelServerList {
		fmt.Println(ui.FormatListItem(modelServer))
	}

	return nil
}

func serversHandler(cmd *cobra.Command, args []string) {
	fmt.Println(ui.FormatHeader("Supported Model Servers", "Install any of these to use with golms"))
	for _, modelServer := range constants.AvailableModelServers {
		fmt.Println(ui.FormatListItem(modelServer))
	}
}

func connectHandler(cmd *cobra.Command, args []string) error {
	// Initialize data objects
	reader := bufio.NewReader(os.Stdin)
	var selectedModelServer string
	var selectedLLM string

	// Get list of all model servers
	modelServerList, err := discovery.ListAllModelServers()
	if err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error encountered while listing model servers: %v", err)))
		return err
	}
	if len(modelServerList) == 0 {
		fmt.Println(ui.FormatWarning("No model servers available"))
		fmt.Println(ui.SubtleStyle.Render("Install one of the following supported model servers:"))
		for _, modelServer := range constants.AvailableModelServers {
			fmt.Println(ui.FormatListItem(modelServer))
		}
		return errors.New("no model servers found")
	}

	// Let user choose a model server with interactive prompt
	modelServerPrompt := promptui.Select{
		Label: ui.PromptStyle.Render("Select a model server"),
		Items: modelServerList,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▸ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: ui.SuccessStyle.Render("✓") + " {{ . }}",
		},
	}

	_, selectedModelServer, err = modelServerPrompt.Run()
	if err != nil {
		fmt.Println(ui.FormatError("Selection cancelled"))
		return err
	}

	fmt.Println()

	// Get following LLMs for that model server
	llmListMap, err := discovery.ListAllLLMs()
	if err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error encountered while listing LLMs: %v", err)))
		return err
	}
	if len(llmListMap) == 0 {
		fmt.Println(ui.FormatWarning("No model server directories found"))
		fmt.Println(ui.SubtleStyle.Render("Make sure models are placed in ~/golms/<model_server>/ directories"))
		return errors.New("no model server directories found")
	}
	llmList, ok := llmListMap[selectedModelServer]
	if !ok {
		fmt.Println(ui.FormatError(fmt.Sprintf("No subdirectory found for model server: %s", selectedModelServer)))
		return fmt.Errorf("no subdirectory found for model server: %s", selectedModelServer)
	}
	if len(llmList) == 0 {
		fmt.Println(ui.FormatError(fmt.Sprintf("No LLMs available for model server: %s", selectedModelServer)))
		return fmt.Errorf("no llms available for model server: %s", selectedModelServer)
	}

	// Let user choose LLM with interactive prompt
	llmPrompt := promptui.Select{
		Label: ui.PromptStyle.Render("Select an LLM"),
		Items: llmList,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▸ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: ui.SuccessStyle.Render("✓") + " {{ . }}",
		},
	}

	_, selectedLLM, err = llmPrompt.Run()
	if err != nil {
		fmt.Println(ui.FormatError("Selection cancelled"))
		return err
	}

	fmt.Println()
	fmt.Println(ui.FormatDivider())

	// Create model server instance
	modelServerManager := server.NewServerManager(selectedModelServer, selectedLLM)
	if !modelServerManager.IsAvailable() {
		fmt.Println(ui.FormatError(fmt.Sprintf("Model server not available: %s", selectedModelServer)))
		return errors.New("model server unavailable")
	}

	running, _ := modelServerManager.IsRunning()
	if !running {
		fmt.Println()
		fmt.Println(ui.HeaderStyle.Render("Starting Model Server"))
		fmt.Println()

		// Show spinner while starting server
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = "  Initializing server...\n"
		s.Start()

		err := modelServerManager.Start()
		s.Stop()

		if err != nil {
			fmt.Println(ui.FormatError(fmt.Sprintf("Failed to start model server: %v", err)))
			return err
		}
		fmt.Println(ui.FormatSuccess("Model server is ready"))
		fmt.Println()
		defer modelServerManager.Stop()
	} else {
		fmt.Println(ui.SubtleStyle.Render("Model server already running"))
		fmt.Println()
	}

	port, err := modelServerManager.GetPort()
	if err != nil {
		fmt.Println(ui.FormatError("Failed to get server port"))
		return err
	}

	fmt.Println(ui.FormatDivider())
	fmt.Println()

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
