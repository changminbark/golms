package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/changminbark/golms/pkg/constants"
	"github.com/changminbark/golms/pkg/discovery"
)

func NewCLI() *cobra.Command {
	// Create root command where user types golms
	rootCmd := &cobra.Command{
		Use: "golms",
		Short: "Local Model Server Interface written in Go",
		SilenceUsage: true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print(cmd.UsageString())
		},
	}

	// Create list command that lists available LLMs and model servers
	listCmd := &cobra.Command {
		Use: "list",
		Short: "List all available LLMs and model servers",
		RunE: listHandler,
	}

	// Create servers command that lists support model servers
	serversCmd := &cobra.Command {
		Use: "servers",
		Short: "List all supported model servers",
		Run: serversHandler,
	}

	// 

	// Add subcommands to root command
	rootCmd.AddCommand(listCmd, serversCmd)

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
		fmt.Printf("Error encountered while listing model servers: %v", err)
		return err
	}
	if len(modelServerList) == 0 {
		fmt.Print("You have no model servers available. \nMake sure to download the following supported models:\n")
		for _, modelServer := range(constants.AvailableModelServers) {
			fmt.Printf("- %s\n", modelServer)
		}
		return errors.New("no model servers found")
	}

	// Print available model servers
	fmt.Print("You have the following model servers available:\n")
	for _, modelServer := range(modelServerList) {
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