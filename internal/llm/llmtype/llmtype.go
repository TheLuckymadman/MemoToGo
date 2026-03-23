package llmtype

import (
	"context"
	"encoding/json"
)

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

type Message struct {
	Role     string    `json:"role"` // "system", "user", "assistant", "tool"
	Content  string    `json:"content"`
	ToolCall *ToolCall `json:"function_call,omitempty"`
	ToolName string    `json:"tool_name,omitempty"`
}

type ToolCall struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	//Arguments map[string]interface{} `json:"arguments"`
	Arguments json.RawMessage `json:"arguments"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens          int `json:"prompt_tokens"`
	CompletionTokens      int `json:"completion_tokens"`
	PrecachedPromptTokens int `json:"precached_prompt_tokens"`
	TotalTokens           int `json:"total_tokens"`
}

type Tool interface {
	Name() string
	Description() string
	Parameters() JSONSchema
	ReturnParameters() JSONSchema
	Call(ctx context.Context, args json.RawMessage) (any, error)
}

type JSONSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}
