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

	listCmd := &cobra.Command {
		Use: "list",
		Short: "List all available LLMs and model servers",
		RunE: listHandler,
	}

	rootCmd.AddCommand(listCmd)

	return rootCmd
}

func listHandler(cmd *cobra.Command, args []string) error {
	// Get list of all LLMs
	llmList, err := discovery.ListAllLLMs()
	if err != nil {
		fmt.Printf("Error encountered while listing LLMS: %v", err)
		return err
	}
	if len(llmList) == 0 {
		fmt.Print("You have no LLMs available in the ~/LLMs directory. \nMake sure to download them as folders with model weights.")
		return errors.New("no LLMs found")
	}

	// Print available LLMs
	fmt.Print("You have the following LLMs available:\n")
	for _, llm := range(llmList) {
		fmt.Printf("- %s\n", llm)
	}

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