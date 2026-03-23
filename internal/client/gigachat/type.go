package gigachat

import "encoding/json"

type ChatRequest struct {
	Model        string       `json:"model"`
	Messages     []ReqMessage `json:"messages"`
	FunctionCall string       `json:"function_call,omitempty"`
	Functions    []Function   `json:"functions,omitempty"`

	Temperature       float64 `json:"temperature,omitempty"`
	TopP              float64 `json:"top_p,omitempty"`
	Stream            bool    `json:"stream,omitempty"`
	MaxTokens         int     `json:"max_tokens,omitempty"`
	RepetitionPenalty float64 `json:"repetition_penalty,omitempty"`
	UpdateInterval    int     `json:"update_interval,omitempty"`
}

type ReqMessage struct {
	Role         string        `json:"role"` // "system", "user", "assistant", "tool"
	Content      string        `json:"content"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

type Function struct {
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	Parameters       JSONSchema       `json:"parameters"`
	FewShotExamples  []FewShotExample `json:"few_shot_examples,omitempty"`
	ReturnParameters JSONSchema       `json:"return_parameters,omitempty"`
}

type FewShotExample struct {
	Request string         `json:"request"`
	Params  map[string]any `json:"params"`
}

type JSONSchema struct {
	Type        string              `json:"type"`
	Properties  map[string]Property `json:"properties,omitempty"`
	Required    []string            `json:"required,omitempty"`
	Items       *JSONSchema         `json:"items,omitempty"` // for arrays
	Enum        []string            `json:"enum,omitempty"`
	Description string              `json:"description,omitempty"`
}

type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Items       *JSONSchema `json:"items,omitempty"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Object  string   `json:"object"`
}

type Choice struct {
	Message      RespMessage `json:"message"`
	Index        int         `json:"index"`
	FinishReason string      `json:"finish_reason"`
}

type RespMessage struct {
	Role             string        `json:"role"`
	Content          string        `json:"content"`
	Created          int64         `json:"created,omitempty"`
	Name             string        `json:"name,omitempty"`
	FunctionsStateID string        `json:"functions_state_id,omitempty"`
	FunctionCall     *FunctionCall `json:"function_call,omitempty"`
}

type FunctionCall struct {
	Name string `json:"name"`
	//Arguments map[string]interface{} `json:"arguments"`
	Arguments json.RawMessage `json:"arguments"`
}

type Usage struct {
	PromptTokens          int `json:"prompt_tokens"`
	CompletionTokens      int `json:"completion_tokens"`
	PrecachedPromptTokens int `json:"precached_prompt_tokens"`
	TotalTokens           int `json:"total_tokens"`
}
