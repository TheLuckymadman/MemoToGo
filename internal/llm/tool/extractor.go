package tool

import "github.com/theluckymadman/memotogo/internal/llm/llmtype"

func ExtractToolCall(resp *llmtype.ChatResponse) *llmtype.ToolCall {
	if resp == nil || len(resp.Choices) == 0 {
		return nil
	}
	choice := resp.Choices[0]

	if tc := choice.Message.ToolCall; tc != nil && tc.Name != "" {
		return tc
	}
	return nil
}
