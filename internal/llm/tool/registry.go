package tool

import (
	"github.com/theluckymadman/memotogo/internal/llm/llmtype"
)

type Registry map[string]llmtype.Tool

func NewRegistry() Registry {
	return make(Registry)
}

func (r Registry) Add(t llmtype.Tool) {
	r[t.Name()] = t
}

func (r Registry) Get(name string) llmtype.Tool {
	return r[name]
}
