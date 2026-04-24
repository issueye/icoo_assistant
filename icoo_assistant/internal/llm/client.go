package llm

import "encoding/json"

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type ToolUse struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

type TextBlock struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text"`
}

type ToolResultBlock struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

type Response struct {
	StopReason string      `json:"stop_reason"`
	Text       string      `json:"text,omitempty"`
	ToolUses   []ToolUse   `json:"tool_uses,omitempty"`
	Raw        interface{} `json:"raw,omitempty"`
}

type Client interface {
	CreateMessage(system string, messages []Message, tools []Tool) (Response, error)
}

type FakeClient struct {
	Responses []Response
	Calls     int
	Snapshots []string
}

func (f *FakeClient) CreateMessage(system string, messages []Message, tools []Tool) (Response, error) {
	f.Calls++
	if data, err := json.Marshal(messages); err == nil {
		f.Snapshots = append(f.Snapshots, string(data))
	}
	if len(f.Responses) == 0 {
		return Response{StopReason: "end", Text: ""}, nil
	}
	resp := f.Responses[0]
	f.Responses = f.Responses[1:]
	return resp, nil
}
