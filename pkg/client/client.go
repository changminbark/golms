package client

import (
	"bufio"
	"errors"

	"github.com/changminbark/golms/pkg/constants"
)

type Message struct {
	Role      string        `json:"role"`
	Content   string        `json:"content"`
	ToolCalls []interface{} `json:"tool_calls"`
}

type ChatOptions struct {
	Temperature float64
	MaxTokens   int
	Stream      bool
}

type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	Stream      bool      `json:"stream"`
}

type ChatResponse struct {
	ID                string   `json:"id"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Object            string   `json:"object"`
	Model             string   `json:"model"`
	Created           int64    `json:"created"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
}

type Choice struct {
	Index        int      `json:"index"`
	FinishReason string   `json:"finish_reason"`
	LogProbs     LogProbs `json:"logprobs"`
	Message      Message  `json:"message"`
}

type LogProbs struct {
	TokenLogProbs []float64     `json:"token_logprobs"`
	TopLogProbs   []interface{} `json:"top_logprobs"` // Empty array in this case
	Tokens        []int         `json:"tokens"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ModelServerClient interface {
	StartChat() error
	setChatOptions()
	addUserMessage(req *ChatRequest) error
	sendChatReq(req *ChatRequest) (*ChatResponse, error)
}

var ErrExitRequested = errors.New("user requested exit")

func NewClient(model_server string, llm string, host string, port int, reader *bufio.Reader) ModelServerClient {
	// Set default ChatOptions for initialization
	defaultChatOptions := ChatOptions{
		Temperature: 0.7,
		MaxTokens:   512,
		Stream:      false,
	}

	// Create client
	switch model_server {
	case constants.Mlx_lm:
		return &MlxLMClient{llm, host, port, defaultChatOptions, reader}
	default:
		return nil
	}
}
