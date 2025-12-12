package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/changminbark/golms/pkg/utils"
)

type MlxLMClient struct {
	llm         string
	host        string
	port        int
	chatOptions ChatOptions
	reader      *bufio.Reader
}

func (c MlxLMClient) StartChat() error {
	// Set Chat Options
	c.setChatOptions()

	fmt.Printf("\n\n======== STARTING CHAT CLIENT WITH mlx_lm model server (%s) ========\n\n", c.llm)
	fmt.Print("Type '/exit' to quit the chat\n\n")

	// Create initial chat request
	chatReq := &ChatRequest{
		Messages:    []Message{},
		Temperature: c.chatOptions.Temperature,
		MaxTokens:   c.chatOptions.MaxTokens,
		Stream:      c.chatOptions.Stream,
	}

	// Create infinite loop for chat
	for {
		// Prompt user for message and add to conversation thread
		err := c.addUserMessage(chatReq)
		if err != nil {
			if errors.Is(err, ErrExitRequested) {
				fmt.Println("\nExiting chat. Goodbye!")
				return nil
			}
			return err
		}

		// Send chat request to model server
		resp, err := c.sendChatReq(chatReq)
		if err != nil {
			return err
		}

		// Clean and display chat response
		cleanedContent := utils.RemoveThinkTags(resp.Choices[0].Message.Content)
		fmt.Printf("LLM(%s): %s\n\n", c.llm, cleanedContent)
	}
}

func (c *MlxLMClient) setChatOptions() {
	fmt.Println("\n=== Chat Options Setup ===")

	// Set Temperature
	fmt.Print("Enter temperature (0.0-2.0, default 0.7): ")
	tempInput, _ := c.reader.ReadString('\n')
	tempInput = strings.TrimSpace(tempInput)

	if tempInput == "" {
		c.chatOptions.Temperature = 0.7
	} else {
		temp, err := strconv.ParseFloat(tempInput, 64)
		if err != nil || temp < 0 || temp > 2.0 {
			fmt.Println("Invalid temperature, using default 0.7")
			c.chatOptions.Temperature = 0.7
		} else {
			c.chatOptions.Temperature = temp
		}
	}

	// Set MaxTokens
	fmt.Print("Enter max tokens (default 512): ")
	tokensInput, _ := c.reader.ReadString('\n')
	tokensInput = strings.TrimSpace(tokensInput)

	if tokensInput == "" {
		c.chatOptions.MaxTokens = 512
	} else {
		tokens, err := strconv.Atoi(tokensInput)
		if err != nil || tokens < 1 {
			fmt.Println("Invalid max tokens, using default 512")
			c.chatOptions.MaxTokens = 512
		} else {
			c.chatOptions.MaxTokens = tokens
		}
	}

	// Set Stream
	fmt.Print("Enable streaming? (y/n, default n): ")
	streamInput, _ := c.reader.ReadString('\n')
	streamInput = strings.TrimSpace(strings.ToLower(streamInput))

	c.chatOptions.Stream = (streamInput == "y" || streamInput == "yes")

	fmt.Println("\n=== Options Set ===")
	fmt.Printf("Temperature: %.2f\n", c.chatOptions.Temperature)
	fmt.Printf("Max Tokens: %d\n", c.chatOptions.MaxTokens)
	fmt.Printf("Streaming: %v\n", c.chatOptions.Stream)
	fmt.Println()
}

func (c *MlxLMClient) addUserMessage(req *ChatRequest) error {
	// Ask user for input
	fmt.Print("User: ")
	userInput, err := c.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	// Trim whitespace and check for exit command
	userInput = strings.TrimSpace(userInput)
	if userInput == "/exit" {
		return ErrExitRequested
	}

	// Create user message
	userMessage := &Message{
		Role:      "user",
		Content:   userInput,
		ToolCalls: nil,
	}

	// Append new user input into request messages
	req.Messages = append(req.Messages, *userMessage)

	return nil
}

func (c *MlxLMClient) sendChatReq(req *ChatRequest) (*ChatResponse, error) {
	// Create data payload of chat request
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Create HTTP Post request
	url := fmt.Sprintf("http://%s:%d/v1/chat/completions", c.host, c.port)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Decode the response
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Add top response to original chat request for conversation context
	req.Messages = append(req.Messages, chatResp.Choices[0].Message)

	return &chatResp, nil
}
