package tools

import (
	"fmt"
	"sort"

	"icoo_assistant/internal/llm"
)

type Registry struct {
	defs map[string]Definition
}

func NewRegistry(defs ...Definition) (*Registry, error) {
	registry := &Registry{defs: map[string]Definition{}}
	for _, def := range defs {
		if def.Tool.Name == "" {
			return nil, fmt.Errorf("tool name required")
		}
		if def.Handler == nil {
			return nil, fmt.Errorf("handler required for tool %s", def.Tool.Name)
		}
		registry.defs[def.Tool.Name] = def
	}
	return registry, nil
}

func (r *Registry) Tools() []llm.Tool {
	tools := make([]llm.Tool, 0, len(r.defs))
	for _, def := range r.defs {
		tools = append(tools, def.Tool)
	}
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	return tools
}

func (r *Registry) Execute(call Call) (Result, error) {
	def, ok := r.defs[call.Name]
	if !ok {
		return Result{}, fmt.Errorf("unknown tool: %s", call.Name)
	}
	content, err := def.Handler(call)
	if err != nil {
		return Result{}, err
	}
	return Result{ToolUseID: call.ID, Type: "tool_result", Content: content}, nil
}
