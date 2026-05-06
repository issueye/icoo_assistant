package tools

import "icoo_assistant/internal/llm"

type Call struct {
	ID    string
	Name  string
	Input map[string]interface{}
}

type Result = llm.ToolResultBlock

type Handler func(Call) (string, error)

type Definition struct {
	Tool    llm.Tool
	Handler Handler
}
